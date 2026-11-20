package agent

import (
	"context"
	"errors"
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

// TestCheckHealthMergesProbeAndKeepsCatalogIndependent 验证同服务探测合并且不触碰目录或 Agent Tool 记录。
func TestCheckHealthMergesProbeAndKeepsCatalogIndependent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:mcp_health?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.MCPServer{}, &model.MCPTool{}, &model.AgentRunTool{}))
	refreshedAt := time.Now().Add(-time.Hour)
	server := model.MCPServer{
		Name: "资源", Code: "resources", BaseURL: "http://resources/mcp", Enabled: true, RequestTimeoutSeconds: 15,
		CatalogStatus: "available", CatalogToolCount: 7, CatalogRefreshedAt: &refreshedAt,
	}
	require.NoError(t, db.Create(&server).Error)

	pingStarted := make(chan struct{})
	releasePing := make(chan struct{})
	client := &healthTestClient{pingStarted: pingStarted, releasePing: releasePing}
	registry := NewMCPRegistry(db)
	var created atomic.Int64
	registry.healthClient = func(model.MCPServer) (mcpHealthClient, error) {
		created.Add(1)
		return client, nil
	}

	errorsFound := make(chan error, 2)
	var calls sync.WaitGroup
	for range 2 {
		calls.Go(func() {
			_, err := registry.CheckHealth(t.Context(), server.Code)
			errorsFound <- err
		})
	}
	<-pingStarted
	require.Eventually(t, func() bool { return created.Load() == 1 }, time.Second, 5*time.Millisecond)
	close(releasePing)
	calls.Wait()
	close(errorsFound)
	for err := range errorsFound {
		require.NoError(t, err)
	}
	require.EqualValues(t, 1, client.starts.Load())
	require.EqualValues(t, 1, client.initializes.Load())
	require.EqualValues(t, 1, client.pings.Load())
	require.EqualValues(t, 1, client.closes.Load())

	var stored model.MCPServer
	require.NoError(t, db.First(&stored, server.ID).Error)
	require.Equal(t, "available", stored.LastStatus)
	require.NotNil(t, stored.LastCheckedAt)
	require.Equal(t, "available", stored.CatalogStatus)
	require.Equal(t, 7, stored.CatalogToolCount)
	require.Equal(t, refreshedAt.Unix(), stored.CatalogRefreshedAt.Unix())
	var toolCalls int64
	require.NoError(t, db.Model(&model.AgentRunTool{}).Count(&toolCalls).Error)
	require.Zero(t, toolCalls)
}

// TestCheckHealthRedactsProbeFailure 验证探测失败只保存安全错误摘要。
func TestCheckHealthRedactsProbeFailure(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.MCPServer{}))
	server := model.MCPServer{Name: "故障", Code: "broken", BaseURL: "http://broken/mcp", AuthToken: "health-secret", Enabled: true, RequestTimeoutSeconds: 15}
	require.NoError(t, db.Create(&server).Error)
	registry := NewMCPRegistry(db)
	registry.healthClient = func(model.MCPServer) (mcpHealthClient, error) {
		return &healthTestClient{pingErr: errors.New("ping failed with health-secret")}, nil
	}

	checked, err := registry.CheckHealth(t.Context(), server.Code)
	require.Error(t, err)
	require.Equal(t, "unavailable", checked.LastStatus)
	require.NotContains(t, checked.LastError, "health-secret")
	require.Contains(t, checked.LastError, "[redacted]")
}

// TestJitteredHealthIntervalStaysWithinTenPercent 验证健康检查周期抖动边界。
func TestJitteredHealthIntervalStaysWithinTenPercent(t *testing.T) {
	for range 100 {
		interval := jitteredHealthInterval(time.Minute)
		require.GreaterOrEqual(t, interval, 54*time.Second)
		require.LessOrEqual(t, interval, 66*time.Second)
	}
}

// healthTestClient 是不暴露 ListTools 的健康探测测试客户端。
type healthTestClient struct {
	starts, initializes, pings, closes atomic.Int64
	pingStarted                        chan struct{}
	releasePing                        chan struct{}
	pingErr                            error
}

func (c *healthTestClient) Start(context.Context) error { c.starts.Add(1); return nil }
func (c *healthTestClient) Initialize(context.Context, mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	c.initializes.Add(1)
	return new(mcp.InitializeResult), nil
}
func (c *healthTestClient) Ping(context.Context) error {
	c.pings.Add(1)
	if c.pingStarted != nil {
		close(c.pingStarted)
	}
	if c.releasePing != nil {
		<-c.releasePing
	}
	return c.pingErr
}
func (c *healthTestClient) Close() error { c.closes.Add(1); return nil }
