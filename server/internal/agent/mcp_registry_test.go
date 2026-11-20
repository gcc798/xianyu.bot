package agent

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/eino-contrib/jsonschema"
	agenttool "github.com/gcc798/xianyu.bot/internal/agent/tool"
	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/glebarez/sqlite"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRunnerToolCatalogIncludesLocalAndAvailableMCPTools(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.MCPServer{}, &model.MCPTool{}))
	require.NoError(t, db.Create(&[]model.MCPServer{
		{Name: "Telegram", Code: "telegram", BaseURL: "http://telegram/mcp", Enabled: true, RequestTimeoutSeconds: 15},
		{Name: "离线服务", Code: "offline", BaseURL: "http://offline/mcp", Enabled: true, RequestTimeoutSeconds: 15},
	}).Error)
	registry := agenttool.NewRegistry()
	require.NoError(t, registry.Register(agenttool.NewCurrentTime()))
	require.NoError(t, registry.Register(&catalogTestTool{code: "xianyu_order_list", name: "查询闲鱼订单"}))
	now := time.Now()
	require.NoError(t, db.Create(&[]model.MCPTool{
		catalogModelTool("telegram", "dialog_list", `{"name":"dialog_list","description":"列出会话","inputSchema":{"type":"object"}}`, now),
		catalogModelTool("telegram", strings.Repeat("a", 120), `{"name":"`+strings.Repeat("a", 120)+`","inputSchema":{"type":"object"}}`, now),
	}).Error)
	mcpRegistry := NewMCPRegistry(db)
	connected := false
	mcpRegistry.connect = func(context.Context, model.MCPServer) (*mcpConnection, error) { connected = true; return nil, nil }
	require.NoError(t, mcpRegistry.LoadCatalog(t.Context()))
	runner := &Runner{db: db, registry: mcpRegistry, tools: registry}

	entries := runner.ToolCatalog(t.Context())
	require.Len(t, entries, 3)
	require.Equal(t, ToolCatalogEntry{Code: "current_time", Name: "获取当前时间", Description: "仅在需要执行当刻的准确时间时调用。日期和星期使用本轮上下文。", Source: "本地", Type: "tool"}, entries[0])
	require.Equal(t, ToolCatalogEntry{Code: "xianyu_order_list", Name: "查询闲鱼订单", Description: "测试工具", Source: "闲鱼", Type: "tool"}, entries[1])
	require.Equal(t, "telegram__dialog_list", entries[2].Code)
	require.Equal(t, "Telegram", entries[2].Source)
	require.Equal(t, "telegram", entries[2].ProviderCode)
	require.False(t, connected)
}

type catalogTestTool struct{ code, name string }

func (t *catalogTestTool) Definition() agenttool.Definition {
	return agenttool.Definition{
		Code: t.code, Name: t.name, Description: "测试工具", Source: "local",
		Parameters: jsonschema.Reflect(struct{}{}), ReadOnly: true, Timeout: time.Second,
	}
}

func (*catalogTestTool) Invoke(context.Context, agenttool.Scope, json.RawMessage) (string, error) {
	return "{}", nil
}

// fakeMCPClient 是测试辅助类型。
type fakeMCPClient struct{}

// Start 提供测试辅助能力。
func (*fakeMCPClient) Start(context.Context) error { return nil }

// Initialize 提供测试辅助能力。
func (*fakeMCPClient) Initialize(context.Context, mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	return new(mcp.InitializeResult), nil
}

// ListTools 提供测试辅助能力。
func (*fakeMCPClient) ListTools(context.Context, mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	return new(mcp.ListToolsResult), nil
}

// CallTool 提供测试辅助能力。
func (*fakeMCPClient) CallTool(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return new(mcp.CallToolResult), nil
}

// Close 提供测试辅助能力。
func (*fakeMCPClient) Close() error { return nil }
