package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/glebarez/sqlite"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestMCPRuntimeSharesClientAndAllowsConcurrentCalls 验证首次建连合并、共享调用并发和 Owner 隔离。
func TestMCPRuntimeSharesClientAndAllowsConcurrentCalls(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:mcp_runtime?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.MCPServer{}))
	server := model.MCPServer{
		Name: "热门服务", Code: "popular", BaseURL: "http://popular/mcp", Enabled: true,
		RequestTimeoutSeconds: 15, ConnectionScope: "shared",
	}
	require.NoError(t, db.Create(&server).Error)

	registry := NewMCPRegistry(db)
	client := new(runtimeTestClient)
	var connections atomic.Int64
	registry.runtimeConnect = func(context.Context, model.MCPServer) (connectedMCPClient, error) {
		connections.Add(1)
		time.Sleep(20 * time.Millisecond)
		return client, nil
	}
	callRuntimeConcurrently(t, registry, 11, server.Code, 10)
	require.EqualValues(t, 1, connections.Load())
	require.EqualValues(t, 10, client.calls.Load())
	require.Greater(t, client.maximum.Load(), int64(1), "共享客户端调用不应被注册表互斥锁串行化")

	require.NoError(t, db.Model(&server).Update("connection_scope", "owner").Error)
	_, err = registry.CallMCPTool(t.Context(), 11, server.Code, mcp.CallToolRequest{})
	require.NoError(t, err)
	_, err = registry.CallMCPTool(t.Context(), 22, server.Code, mcp.CallToolRequest{})
	require.NoError(t, err)
	require.EqualValues(t, 3, connections.Load())
	require.Len(t, registry.runtimeClients, 3)
}

// TestMCPRuntimeSharesFirstConnectionFailure 验证并发首次调用共享同一次失败。
func TestMCPRuntimeSharesFirstConnectionFailure(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:mcp_runtime_failure?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.MCPServer{}))
	server := model.MCPServer{Name: "故障", Code: "broken", BaseURL: "http://broken/mcp", Enabled: true, ConnectionScope: "shared"}
	require.NoError(t, db.Create(&server).Error)
	registry := NewMCPRegistry(db)
	var connections atomic.Int64
	registry.runtimeConnect = func(context.Context, model.MCPServer) (connectedMCPClient, error) {
		connections.Add(1)
		time.Sleep(20 * time.Millisecond)
		return nil, errors.New("connection refused")
	}

	errorsFound := make(chan error, 5)
	var calls sync.WaitGroup
	for range 5 {
		calls.Go(func() {
			_, err := registry.CallMCPTool(t.Context(), 1, server.Code, mcp.CallToolRequest{})
			errorsFound <- err
		})
	}
	calls.Wait()
	close(errorsFound)
	for err := range errorsFound {
		require.ErrorContains(t, err, "connection refused")
	}
	require.EqualValues(t, 1, connections.Load())
}

// TestMCPRuntimeConcurrencySlotHonorsContext 验证服务级并发等待遵守调用 Context。
func TestMCPRuntimeConcurrencySlotHonorsContext(t *testing.T) {
	db, server := runtimeTestDB(t, "limited", 1)
	registry := NewMCPRegistry(db)
	started := make(chan struct{})
	release := make(chan struct{})
	client := &runtimeOutcomeClient{started: started, release: release}
	registry.runtimeConnect = func(context.Context, model.MCPServer) (connectedMCPClient, error) { return client, nil }

	firstDone := make(chan error, 1)
	go func() {
		_, err := registry.CallMCPTool(t.Context(), 1, server.Code, mcp.CallToolRequest{})
		firstDone <- err
	}()
	<-started
	waitContext, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()
	_, err := registry.CallMCPTool(waitContext, 2, server.Code, mcp.CallToolRequest{})
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.EqualValues(t, 1, client.calls.Load())

	registry.runtimeMu.Lock()
	connection := registry.runtimeClients[server.Code]
	registry.runtimeMu.Unlock()
	connection.mu.Lock()
	require.Equal(t, 1, connection.inflight)
	connection.mu.Unlock()
	close(release)
	require.NoError(t, <-firstDone)
}

// TestMCPRuntimeTransportFailureInvalidatesGenerationWithoutRetry 验证传输失败不重放并使当前代次失效。
func TestMCPRuntimeTransportFailureInvalidatesGenerationWithoutRetry(t *testing.T) {
	db, server := runtimeTestDB(t, "unstable", 0)
	registry := NewMCPRegistry(db)
	first := &runtimeOutcomeClient{callErr: errors.New("connection reset")}
	second := &runtimeOutcomeClient{result: new(mcp.CallToolResult)}
	clients := []*runtimeOutcomeClient{first, second}
	var connections atomic.Int64
	registry.runtimeConnect = func(context.Context, model.MCPServer) (connectedMCPClient, error) {
		index := int(connections.Add(1)) - 1
		return clients[index], nil
	}

	_, err := registry.CallMCPTool(t.Context(), 1, server.Code, mcp.CallToolRequest{})
	require.ErrorContains(t, err, "connection reset")
	require.EqualValues(t, 1, first.calls.Load(), "结果未知的调用不得自动重放")
	require.EqualValues(t, 1, first.closes.Load())
	require.Empty(t, registry.runtimeClients)
	var stored model.MCPServer
	require.NoError(t, db.First(&stored, server.ID).Error)
	require.Equal(t, "unavailable", stored.LastStatus)

	_, err = registry.CallMCPTool(t.Context(), 1, server.Code, mcp.CallToolRequest{})
	require.NoError(t, err)
	require.EqualValues(t, 2, connections.Load())
	require.Len(t, registry.runtimeClients, 1)
}

// TestMCPRuntimeBusinessErrorKeepsConnectionAvailable 验证 MCP 业务错误不被误判为离线。
func TestMCPRuntimeBusinessErrorKeepsConnectionAvailable(t *testing.T) {
	db, server := runtimeTestDB(t, "business", 0)
	registry := NewMCPRegistry(db)
	client := &runtimeOutcomeClient{result: &mcp.CallToolResult{IsError: true}}
	var connections atomic.Int64
	registry.runtimeConnect = func(context.Context, model.MCPServer) (connectedMCPClient, error) {
		connections.Add(1)
		return client, nil
	}

	result, err := registry.CallMCPTool(t.Context(), 1, server.Code, mcp.CallToolRequest{})
	require.NoError(t, err)
	require.True(t, result.IsError)
	_, err = registry.CallMCPTool(t.Context(), 1, server.Code, mcp.CallToolRequest{})
	require.NoError(t, err)
	require.EqualValues(t, 1, connections.Load())
	require.Zero(t, client.closes.Load())
	var stored model.MCPServer
	require.NoError(t, db.First(&stored, server.ID).Error)
	require.Equal(t, "available", stored.LastStatus)
}

// TestMCPRuntimeKeepsAnyNumberAndSweepsOnlyIdleConnections 验证无全局上限、TTL 回收及健康/目录读取不续期。
func TestMCPRuntimeKeepsAnyNumberAndSweepsOnlyIdleConnections(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:mcp_runtime_many?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.MCPServer{}, &model.MCPTool{}))
	registry := NewMCPRegistry(db)
	clients := make([]*runtimeOutcomeClient, 0, 30)
	registry.runtimeConnect = func(context.Context, model.MCPServer) (connectedMCPClient, error) {
		client := new(runtimeOutcomeClient)
		clients = append(clients, client)
		return client, nil
	}
	for i := range 30 {
		code := fmt.Sprintf("bulk_%d", i)
		server := model.MCPServer{
			Name: code, Code: code, BaseURL: "http://" + code + "/mcp", Enabled: true,
			RequestTimeoutSeconds: 15, ConnectionScope: "shared", ConnectionTTLSeconds: 10,
		}
		require.NoError(t, db.Create(&server).Error)
		_, err := registry.CallMCPTool(t.Context(), 1, code, mcp.CallToolRequest{})
		require.NoError(t, err)
	}
	require.Len(t, registry.runtimeClients, 30)

	connection := registry.runtimeClients["bulk_0"]
	connection.mu.Lock()
	lastUsedAt := connection.lastUsedAt
	connection.mu.Unlock()
	registry.healthClient = func(model.MCPServer) (mcpHealthClient, error) { return new(healthTestClient), nil }
	_, err = registry.CheckHealth(t.Context(), "bulk_0")
	require.NoError(t, err)
	_ = registry.CatalogTools()
	connection.mu.Lock()
	require.True(t, connection.lastUsedAt.Equal(lastUsedAt))
	connection.mu.Unlock()

	registry.sweepIdleRuntime(lastUsedAt.Add(time.Hour))
	require.Empty(t, registry.runtimeClients)
	for _, client := range clients {
		require.EqualValues(t, 1, client.closes.Load())
	}
}

// TestMCPRuntimeSweepAndShutdownWaitForInflight 验证 TTL 和退出清理都不会关闭在途连接。
func TestMCPRuntimeSweepAndShutdownWaitForInflight(t *testing.T) {
	db, server := runtimeTestDB(t, "inflight", 0)
	registry := NewMCPRegistry(db)
	started := make(chan struct{})
	release := make(chan struct{})
	client := &runtimeOutcomeClient{started: started, release: release}
	registry.runtimeConnect = func(context.Context, model.MCPServer) (connectedMCPClient, error) { return client, nil }
	callDone := make(chan error, 1)
	go func() {
		_, err := registry.CallMCPTool(t.Context(), 1, server.Code, mcp.CallToolRequest{})
		callDone <- err
	}()
	<-started
	registry.sweepIdleRuntime(time.Now().Add(time.Hour))
	require.Len(t, registry.runtimeClients, 1)
	require.Zero(t, client.closes.Load())

	closeDone := make(chan error, 1)
	go func() {
		closeContext, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		closeDone <- registry.CloseRuntimeConnections(closeContext)
	}()
	time.Sleep(20 * time.Millisecond)
	require.Zero(t, client.closes.Load())
	close(release)
	require.NoError(t, <-callDone)
	require.NoError(t, <-closeDone)
	require.EqualValues(t, 1, client.closes.Load())
	require.Empty(t, registry.runtimeClients)
}

// callRuntimeConcurrently 并发调用一个 MCP 工具并验证全部成功。
func callRuntimeConcurrently(t *testing.T, registry *MCPRegistry, ownerID uint, serverCode string, count int) {
	t.Helper()
	errorsFound := make(chan error, count)
	var calls sync.WaitGroup
	for range count {
		calls.Go(func() {
			_, err := registry.CallMCPTool(t.Context(), ownerID, serverCode, mcp.CallToolRequest{})
			errorsFound <- err
		})
	}
	calls.Wait()
	close(errorsFound)
	for err := range errorsFound {
		require.NoError(t, err)
	}
}

// runtimeTestClient 记录同一逻辑客户端的并发调用。
type runtimeTestClient struct {
	calls, active, maximum atomic.Int64
}

func (*runtimeTestClient) Start(context.Context) error { return nil }
func (*runtimeTestClient) Initialize(context.Context, mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	return new(mcp.InitializeResult), nil
}
func (*runtimeTestClient) ListTools(context.Context, mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	return new(mcp.ListToolsResult), nil
}
func (c *runtimeTestClient) CallTool(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	c.calls.Add(1)
	current := c.active.Add(1)
	for {
		previous := c.maximum.Load()
		if current <= previous || c.maximum.CompareAndSwap(previous, current) {
			break
		}
	}
	time.Sleep(10 * time.Millisecond)
	c.active.Add(-1)
	return new(mcp.CallToolResult), nil
}
func (*runtimeTestClient) Close() error { return nil }

// runtimeOutcomeClient 返回预设调用结果并可阻塞在途调用。
type runtimeOutcomeClient struct {
	calls, closes atomic.Int64
	started       chan struct{}
	release       chan struct{}
	result        *mcp.CallToolResult
	callErr       error
	startOnce     sync.Once
}

func (*runtimeOutcomeClient) Start(context.Context) error { return nil }
func (*runtimeOutcomeClient) Initialize(context.Context, mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	return new(mcp.InitializeResult), nil
}
func (*runtimeOutcomeClient) ListTools(context.Context, mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	return new(mcp.ListToolsResult), nil
}
func (c *runtimeOutcomeClient) CallTool(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	c.calls.Add(1)
	if c.started != nil {
		c.startOnce.Do(func() { close(c.started) })
	}
	if c.release != nil {
		<-c.release
	}
	if c.result == nil && c.callErr == nil {
		return new(mcp.CallToolResult), nil
	}
	return c.result, c.callErr
}
func (c *runtimeOutcomeClient) Close() error { c.closes.Add(1); return nil }

// runtimeTestDB 创建一个执行连接测试服务。
func runtimeTestDB(t *testing.T, code string, maxConcurrentCalls int) (*gorm.DB, model.MCPServer) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+code+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.MCPServer{}))
	server := model.MCPServer{
		Name: code, Code: code, BaseURL: "http://" + code + "/mcp", Enabled: true,
		RequestTimeoutSeconds: 15, ConnectionScope: "shared", MaxConcurrentCalls: maxConcurrentCalls,
	}
	require.NoError(t, db.Create(&server).Error)
	return db, server
}
