package agent

import (
	"context"
	"encoding/json"
	"errors"
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

// TestMCPRegistryDiscoversEnabledServersAndKeepsAvailableOnes 验证对应测试场景。
func TestMCPRegistryDiscoversEnabledServersAndKeepsAvailableOnes(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.MCPServer{}))
	require.NoError(t, db.Create(&[]model.MCPServer{
		{Name: "可用", Code: "available", BaseURL: "http://available/mcp", Enabled: true, RequestTimeoutSeconds: 15},
		{Name: "故障", Code: "broken", BaseURL: "http://broken/mcp", Enabled: true, RequestTimeoutSeconds: 15},
		{Name: "停用", Code: "disabled", BaseURL: "http://disabled/mcp", Enabled: true, RequestTimeoutSeconds: 15},
	}).Error)
	require.NoError(t, db.Model(&model.MCPServer{}).Where("code = ?", "disabled").Update("enabled", false).Error)

	discovered := make([]string, 0)
	registry := &MCPRegistry{db: db, connect: func(_ context.Context, server model.MCPServer) (*mcpConnection, error) {
		discovered = append(discovered, server.Code)
		if server.Code == "broken" {
			return nil, errors.New("connection refused")
		}
		return &mcpConnection{Server: server, Client: new(fakeMCPClient), Tools: []mcp.Tool{{Name: "resource_search"}}}, nil
	}}
	connections, err := registry.Discover(t.Context())
	require.NoError(t, err)
	require.Equal(t, []string{"available", "broken"}, discovered)
	require.Len(t, connections, 1)
	require.Equal(t, "resource_search", connections[0].Tools[0].Name)

	var available, broken model.MCPServer
	require.NoError(t, db.Where("code = ?", "available").First(&available).Error)
	require.NoError(t, db.Where("code = ?", "broken").First(&broken).Error)
	require.Equal(t, "available", available.LastStatus)
	require.Equal(t, "unavailable", broken.LastStatus)
	require.NotEmpty(t, broken.LastError)
}

func TestRunnerToolCatalogIncludesLocalAndAvailableMCPTools(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.MCPServer{}))
	require.NoError(t, db.Create(&[]model.MCPServer{
		{Name: "Telegram", Code: "telegram", BaseURL: "http://telegram/mcp", Enabled: true, RequestTimeoutSeconds: 15},
		{Name: "离线服务", Code: "offline", BaseURL: "http://offline/mcp", Enabled: true, RequestTimeoutSeconds: 15},
	}).Error)
	registry := agenttool.NewRegistry()
	require.NoError(t, registry.Register(agenttool.NewCurrentTime()))
	require.NoError(t, registry.Register(&catalogTestTool{code: "xianyu_order_list", name: "查询闲鱼订单"}))
	mcpRegistry := &MCPRegistry{db: db, connect: func(_ context.Context, server model.MCPServer) (*mcpConnection, error) {
		if server.Code == "offline" {
			return nil, errors.New("connection refused")
		}
		return &mcpConnection{
			Server: server, Client: new(fakeMCPClient),
			Tools: []mcp.Tool{
				{Name: "dialog_list", Description: "列出会话", InputSchema: mcp.ToolInputSchema{Type: "object"}},
				{Name: strings.Repeat("a", 120), Description: "超长编码", InputSchema: mcp.ToolInputSchema{Type: "object"}},
			},
		}, nil
	}}
	runner := &Runner{db: db, registry: mcpRegistry, tools: registry}

	entries := runner.ToolCatalog(t.Context())
	require.Len(t, entries, 3)
	require.Equal(t, ToolCatalogEntry{Code: "current_time", Name: "获取当前时间", Description: "获取服务端当前日期、时间、星期和时区。回答今天、明天、当前时间、星期几等依赖当前时间的问题时必须调用，不得猜测。", Source: "本地", Type: "tool"}, entries[0])
	require.Equal(t, ToolCatalogEntry{Code: "xianyu_order_list", Name: "查询闲鱼订单", Description: "测试工具", Source: "闲鱼", Type: "tool"}, entries[1])
	require.Equal(t, "telegram__dialog_list", entries[2].Code)
	require.Equal(t, "Telegram", entries[2].Source)
	require.Equal(t, "telegram", entries[2].ProviderCode)
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

// TestMCPRegistryFailsWhenNoEnabledServerIsAvailable 验证对应测试场景。
func TestMCPRegistryFailsWhenNoEnabledServerIsAvailable(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.MCPServer{}))
	require.NoError(t, db.Create(&model.MCPServer{Name: "故障", Code: "broken", BaseURL: "http://broken/mcp", Enabled: true, RequestTimeoutSeconds: 15}).Error)

	registry := &MCPRegistry{db: db, connect: func(context.Context, model.MCPServer) (*mcpConnection, error) {
		return nil, errors.New("connection refused")
	}}
	connections, err := registry.Discover(t.Context())
	require.Nil(t, connections)
	require.ErrorContains(t, err, "MCP broken")
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
