package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/mark3labs/mcp-go/mcp"
)

// mcpRuntimeConnection 保存一个连接作用域内共享的 MCP 执行客户端。
type mcpRuntimeConnection struct {
	server     model.MCPServer
	client     connectedMCPClient
	key        string
	generation uint64
	mu         sync.Mutex
	inflight   int
	lastUsedAt time.Time
	closing    bool
	slots      chan struct{}
	closeOnce  sync.Once
}

// mcpRuntimeConnect 保存可被并发首次调用共享的建连结果。
type mcpRuntimeConnect struct {
	done       chan struct{}
	connection *mcpRuntimeConnection
	err        error
}

// CallMCPTool 按服务连接作用域惰性获取客户端并并发调用远端工具。
func (r *MCPRegistry) CallMCPTool(ctx context.Context, ownerID uint, serverCode string, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var server model.MCPServer
	if err := r.db.WithContext(ctx).Where("code = ? AND enabled = ?", serverCode, true).First(&server).Error; err != nil {
		return nil, err
	}
	connection, err := r.runtimeClient(ctx, ownerID, server)
	if err != nil {
		return nil, err
	}
	if err := connection.acquire(ctx); err != nil {
		return nil, err
	}
	startedAt := time.Now()
	result, callErr := connection.client.CallTool(ctx, request)
	connection.release()
	if callErr != nil {
		r.invalidateRuntime(connection)
		r.updateRuntimeStatus(ctx, server, "unavailable", time.Since(startedAt), callErr)
		return nil, callErr
	}
	r.updateRuntimeStatus(ctx, server, "available", time.Since(startedAt), nil)
	return result, nil
}

// runtimeClient 返回已有客户端或合并同作用域的首次建连。
func (r *MCPRegistry) runtimeClient(ctx context.Context, ownerID uint, server model.MCPServer) (*mcpRuntimeConnection, error) {
	key, err := mcpRuntimeKey(server, ownerID)
	if err != nil {
		return nil, err
	}
	r.runtimeMu.Lock()
	if r.runtimeClosing {
		r.runtimeMu.Unlock()
		return nil, errors.New("MCP 执行连接正在关闭")
	}
	if r.runtimeClients == nil {
		r.runtimeClients = make(map[string]*mcpRuntimeConnection)
		r.runtimeConnecting = make(map[string]*mcpRuntimeConnect)
	}
	if connection := r.runtimeClients[key]; connection != nil {
		r.runtimeMu.Unlock()
		return connection, nil
	}
	if current := r.runtimeConnecting[key]; current != nil {
		r.runtimeMu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-current.done:
			return current.connection, current.err
		}
	}
	current := &mcpRuntimeConnect{done: make(chan struct{})}
	r.runtimeConnecting[key] = current
	r.runtimeMu.Unlock()

	connect := r.runtimeConnect
	if connect == nil {
		connect = connectRuntimeMCPServer
	}
	client, connectErr := connect(ctx, server)
	if connectErr == nil {
		connection := &mcpRuntimeConnection{
			server: server, client: client, key: key, generation: r.runtimeGeneration.Add(1), lastUsedAt: time.Now(),
		}
		if server.MaxConcurrentCalls > 0 {
			connection.slots = make(chan struct{}, server.MaxConcurrentCalls)
		}
		current.connection = connection
	}
	current.err = connectErr
	r.runtimeMu.Lock()
	delete(r.runtimeConnecting, key)
	var rejected *mcpRuntimeConnection
	if r.runtimeClosing && current.connection != nil {
		current.connection.closing = true
		rejected = current.connection
		current.connection = nil
		current.err = errors.New("MCP 执行连接正在关闭")
	} else if current.connection != nil {
		r.runtimeClients[key] = current.connection
	}
	close(current.done)
	r.runtimeMu.Unlock()
	if rejected != nil {
		rejected.close()
	}
	return current.connection, current.err
}

// acquire 获取服务级调用名额和在途租约。
func (c *mcpRuntimeConnection) acquire(ctx context.Context) error {
	if c.slots != nil {
		select {
		case c.slots <- struct{}{}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	c.mu.Lock()
	if c.closing {
		c.mu.Unlock()
		if c.slots != nil {
			<-c.slots
		}
		return errors.New("MCP 执行连接已失效")
	}
	c.inflight++
	c.mu.Unlock()
	return nil
}

// release 释放在途租约，并在失效连接最后一个调用退出时关闭客户端。
func (c *mcpRuntimeConnection) release() {
	c.mu.Lock()
	c.inflight--
	c.lastUsedAt = time.Now()
	shouldClose := c.closing && c.inflight == 0
	c.mu.Unlock()
	if c.slots != nil {
		<-c.slots
	}
	if shouldClose {
		c.close()
	}
}

// close 保证底层客户端最多关闭一次。
func (c *mcpRuntimeConnection) close() { c.closeOnce.Do(func() { _ = c.client.Close() }) }

// invalidateRuntime 仅移除发生连接错误的当前代次，不自动重放工具请求。
func (r *MCPRegistry) invalidateRuntime(connection *mcpRuntimeConnection) {
	r.runtimeMu.Lock()
	if current := r.runtimeClients[connection.key]; current == connection && current.generation == connection.generation {
		delete(r.runtimeClients, connection.key)
	}
	r.runtimeMu.Unlock()
	connection.mu.Lock()
	connection.closing = true
	shouldClose := connection.inflight == 0
	connection.mu.Unlock()
	if shouldClose {
		connection.close()
	}
}

// updateRuntimeStatus 根据传输结果更新健康投影，业务 IsError 不会进入失败路径。
func (r *MCPRegistry) updateRuntimeStatus(ctx context.Context, server model.MCPServer, status string, duration time.Duration, cause error) {
	_ = r.db.WithContext(context.WithoutCancel(ctx)).Model(&model.MCPServer{}).Where("id = ?", server.ID).Updates(map[string]any{
		"last_status": status, "last_checked_at": time.Now(), "last_check_duration_millis": duration.Milliseconds(),
		"last_error": safeMCPError(server, cause),
	}).Error
}

// StartRuntimeCleanup 启动执行连接空闲 TTL 扫描。
func (r *MCPRegistry) StartRuntimeCleanup(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				r.sweepIdleRuntime(now)
			}
		}
	}()
}

// sweepIdleRuntime 原子移除无在途调用且超过空闲 TTL 的连接。
func (r *MCPRegistry) sweepIdleRuntime(now time.Time) {
	toClose := make([]*mcpRuntimeConnection, 0)
	r.runtimeMu.Lock()
	for key, connection := range r.runtimeClients {
		connection.mu.Lock()
		ttl := time.Duration(connection.server.ConnectionTTLSeconds) * time.Second
		if ttl <= 0 {
			ttl = 10 * time.Minute
		}
		if !connection.closing && connection.inflight == 0 && now.Sub(connection.lastUsedAt) >= ttl {
			connection.closing = true
			if current := r.runtimeClients[key]; current == connection && current.generation == connection.generation {
				delete(r.runtimeClients, key)
				toClose = append(toClose, connection)
			}
		}
		connection.mu.Unlock()
	}
	r.runtimeMu.Unlock()
	for _, connection := range toClose {
		connection.close()
	}
}

// CloseRuntimeConnections 停止新租约，等待在途调用后关闭全部执行客户端。
func (r *MCPRegistry) CloseRuntimeConnections(ctx context.Context) error {
	r.runtimeMu.Lock()
	r.runtimeClosing = true
	connections := make([]*mcpRuntimeConnection, 0, len(r.runtimeClients))
	for key, connection := range r.runtimeClients {
		connection.mu.Lock()
		connection.closing = true
		connection.mu.Unlock()
		delete(r.runtimeClients, key)
		connections = append(connections, connection)
	}
	r.runtimeMu.Unlock()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		allIdle := true
		for _, connection := range connections {
			connection.mu.Lock()
			allIdle = allIdle && connection.inflight == 0
			connection.mu.Unlock()
		}
		if allIdle {
			for _, connection := range connections {
				connection.close()
			}
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// mcpRuntimeKey 根据 shared/owner 配置生成不含凭据的连接键。
func mcpRuntimeKey(server model.MCPServer, ownerID uint) (string, error) {
	if server.ConnectionScope != "owner" {
		return server.Code, nil
	}
	if ownerID == 0 {
		return "", errors.New("owner 作用域 MCP 调用缺少用户标识")
	}
	return fmt.Sprintf("%s:%d", server.Code, ownerID), nil
}

// connectRuntimeMCPServer 只初始化执行客户端，不读取工具目录。
func connectRuntimeMCPServer(ctx context.Context, server model.MCPServer) (connectedMCPClient, error) {
	client, err := newMCPHTTPClient(server)
	if err != nil {
		return nil, err
	}
	timeout := time.Duration(server.RequestTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	connectContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := client.Start(connectContext); err != nil {
		_ = client.Close()
		return nil, err
	}
	_, err = client.Initialize(connectContext, mcp.InitializeRequest{Params: mcp.InitializeParams{
		ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
		ClientInfo:      mcp.Implementation{Name: "xy-agent-runtime", Version: "1.0.0"},
	}})
	if err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}
