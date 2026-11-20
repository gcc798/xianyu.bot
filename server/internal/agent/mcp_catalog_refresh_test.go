package agent

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/glebarez/sqlite"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestRefreshCatalogAtomicallyReplacesCompleteToolList 验证完整分页结果、坏工具隔离和目录哈希。
func TestRefreshCatalogAtomicallyReplacesCompleteToolList(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.MCPServer{}, &model.MCPTool{}))
	server := model.MCPServer{Name: "资源", Code: "resources", BaseURL: "http://resources/mcp", Enabled: true, RequestTimeoutSeconds: 15}
	require.NoError(t, db.Create(&server).Error)
	old := newCatalogRecord("resources", "old")
	require.NoError(t, db.Create(&old).Error)

	client := new(refreshMCPClient)
	registry := NewMCPRegistry(db)
	require.NoError(t, registry.LoadCatalog(t.Context()))
	registry.connect = func(_ context.Context, current model.MCPServer) (*mcpConnection, error) {
		return &mcpConnection{Server: current, Client: client, Instructions: "仅用于外部提示", Tools: []mcp.Tool{
			{Name: "page_one", Description: "第一页", InputSchema: mcp.ToolInputSchema{Type: "object"}},
			{Name: "page_two", Description: "第二页", InputSchema: mcp.ToolInputSchema{Type: "object"}},
			{Name: "broken", RawInputSchema: []byte(`{`)},
			{Name: "page_two", Description: "重复", InputSchema: mcp.ToolInputSchema{Type: "object"}},
		}}, nil
	}

	require.NoError(t, registry.RefreshCatalog(t.Context(), server.Code))
	require.True(t, client.closed)
	require.Len(t, registry.CatalogTools(), 2)
	_, oldExists := registry.CatalogTool("resources__old")
	require.False(t, oldExists)
	first, exists := registry.CatalogTool("resources__page_one")
	require.True(t, exists)
	require.Len(t, first.Snapshot.DefinitionHash, 64)

	var storedServer model.MCPServer
	require.NoError(t, db.First(&storedServer, server.ID).Error)
	require.Equal(t, "available", storedServer.CatalogStatus)
	require.Equal(t, 2, storedServer.CatalogToolCount)
	require.Equal(t, "仅用于外部提示", storedServer.ServerInstructions)
	require.Contains(t, storedServer.CatalogError, "broken")
	require.Contains(t, storedServer.CatalogError, "重复")

	var stored []model.MCPTool
	require.NoError(t, db.Order("name").Find(&stored).Error)
	require.Equal(t, []string{"page_one", "page_two"}, []string{stored[0].Name, stored[1].Name})
	require.JSONEq(t, `{"type":"object"}`, stored[0].InputSchema)
}

// TestRefreshCatalogFailureKeepsLastSnapshotAndRedactsError 验证刷新失败保留旧目录并安全截断错误。
func TestRefreshCatalogFailureKeepsLastSnapshotAndRedactsError(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.MCPServer{}, &model.MCPTool{}))
	server := model.MCPServer{
		Name: "资源", Code: "resources", BaseURL: "http://user:password@resources/mcp?token=secret", AuthToken: "secret-token",
		Enabled: true, RequestTimeoutSeconds: 15,
	}
	require.NoError(t, db.Create(&server).Error)
	old := newCatalogRecord("resources", "old")
	require.NoError(t, db.Create(&old).Error)
	registry := NewMCPRegistry(db)
	require.NoError(t, registry.LoadCatalog(t.Context()))
	registry.connect = func(context.Context, model.MCPServer) (*mcpConnection, error) {
		return nil, errors.New("secret-token at http://user:password@resources/mcp?token=secret " + strings.Repeat("x", 1200))
	}

	require.Error(t, registry.RefreshCatalog(t.Context(), server.Code))
	_, exists := registry.CatalogTool("resources__old")
	require.True(t, exists)
	var count int64
	require.NoError(t, db.Model(&model.MCPTool{}).Where("mcp_server_code = ?", server.Code).Count(&count).Error)
	require.EqualValues(t, 1, count)
	var storedServer model.MCPServer
	require.NoError(t, db.First(&storedServer, server.ID).Error)
	require.Equal(t, "unavailable", storedServer.CatalogStatus)
	require.NotContains(t, storedServer.CatalogError, "secret-token")
	require.NotContains(t, storedServer.CatalogError, "user:password")
	require.LessOrEqual(t, len([]rune(storedServer.CatalogError)), 1000)
}

// refreshMCPClient 仅记录刷新短连接是否关闭。
type refreshMCPClient struct{ closed bool }

func (*refreshMCPClient) Start(context.Context) error { return nil }
func (*refreshMCPClient) Initialize(context.Context, mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	return new(mcp.InitializeResult), nil
}
func (*refreshMCPClient) ListTools(context.Context, mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	return new(mcp.ListToolsResult), nil
}
func (*refreshMCPClient) CallTool(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return new(mcp.CallToolResult), nil
}
func (c *refreshMCPClient) Close() error { c.closed = true; return nil }

// newCatalogRecord 创建可恢复的目录记录。
func newCatalogRecord(serverCode, name string) model.MCPTool {
	definition := `{"name":"` + name + `","inputSchema":{"type":"object"}}`
	return model.MCPTool{
		MCPServerCode: serverCode, Name: name, InputSchema: `{"type":"object"}`, Definition: definition,
		SearchText: name, DefinitionHash: name + "-hash", DiscoveredAt: time.Now(),
	}
}
