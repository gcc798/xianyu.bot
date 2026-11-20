package agent

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/glebarez/sqlite"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestRefreshDueCatalogsBoundsConcurrencyAndIsolatesFailure 验证后台刷新并发有界且单服务失败不影响其他服务。
func TestRefreshDueCatalogsBoundsConcurrencyAndIsolatesFailure(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:catalog_scheduler?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&model.MCPServer{}, &model.MCPTool{}))
	for _, code := range []string{"first", "broken", "second"} {
		server := model.MCPServer{Name: code, Code: code, BaseURL: "http://" + code + "/mcp", Enabled: true, RequestTimeoutSeconds: 15, CatalogTTLSeconds: 3600}
		require.NoError(t, db.Create(&server).Error)
	}

	registry := NewMCPRegistry(db)
	var calls, active, maximum atomic.Int64
	registry.connect = func(_ context.Context, server model.MCPServer) (*mcpConnection, error) {
		calls.Add(1)
		current := active.Add(1)
		for {
			previous := maximum.Load()
			if current <= previous || maximum.CompareAndSwap(previous, current) {
				break
			}
		}
		defer active.Add(-1)
		time.Sleep(20 * time.Millisecond)
		if server.Code == "broken" {
			return nil, errors.New("connection refused")
		}
		return &mcpConnection{Server: server, Client: new(refreshMCPClient), Tools: []mcp.Tool{
			{Name: "search", InputSchema: mcp.ToolInputSchema{Type: "object"}},
		}}, nil
	}

	require.Error(t, registry.refreshDueCatalogs(t.Context(), 2))
	require.EqualValues(t, 3, calls.Load())
	require.LessOrEqual(t, maximum.Load(), int64(2))
	require.Len(t, registry.CatalogTools(), 2)
	require.Error(t, registry.refreshDueCatalogs(t.Context(), 2))
	require.EqualValues(t, 4, calls.Load(), "TTL 内只重试上次失败的服务")
	_ = registry.CatalogTools()
	require.EqualValues(t, 4, calls.Load())

	var available int64
	require.NoError(t, db.Model(&model.MCPServer{}).Where("catalog_status = ?", "available").Count(&available).Error)
	require.EqualValues(t, 2, available)
}
