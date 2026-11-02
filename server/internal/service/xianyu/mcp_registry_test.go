package service

import (
	"context"
	"errors"
	"testing"

	"github.com/gcc798/xianyu.bot/internal/infra"
	"github.com/gcc798/xianyu.bot/internal/model/xianyu"
	"github.com/glebarez/sqlite"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMCPRegistryDiscoversEnabledServersAndKeepsAvailableOnes(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.MCPServer{}))
	oldDB := infra.DB
	infra.DB = db
	t.Cleanup(func() { infra.DB = oldDB })

	require.NoError(t, db.Create(&[]model.MCPServer{
		{Name: "可用", Code: "available", BaseURL: "http://available/mcp", Enabled: true, RequestTimeoutSeconds: 15},
		{Name: "故障", Code: "broken", BaseURL: "http://broken/mcp", Enabled: true, RequestTimeoutSeconds: 15},
		{Name: "停用", Code: "disabled", BaseURL: "http://disabled/mcp", Enabled: true, RequestTimeoutSeconds: 15},
	}).Error)
	require.NoError(t, db.Model(&model.MCPServer{}).Where("code = ?", "disabled").Update("enabled", false).Error)

	discovered := make([]string, 0)
	registry := &mcpRegistry{connect: func(_ context.Context, server model.MCPServer) (*mcpConnection, error) {
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

func TestMCPRegistryFailsWhenNoEnabledServerIsAvailable(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.MCPServer{}))
	oldDB := infra.DB
	infra.DB = db
	t.Cleanup(func() { infra.DB = oldDB })
	require.NoError(t, db.Create(&model.MCPServer{Name: "故障", Code: "broken", BaseURL: "http://broken/mcp", Enabled: true, RequestTimeoutSeconds: 15}).Error)

	registry := &mcpRegistry{connect: func(context.Context, model.MCPServer) (*mcpConnection, error) {
		return nil, errors.New("connection refused")
	}}
	connections, err := registry.Discover(t.Context())
	require.Nil(t, connections)
	require.ErrorContains(t, err, "MCP broken")
}

type fakeMCPClient struct{}

func (*fakeMCPClient) Start(context.Context) error { return nil }
func (*fakeMCPClient) Initialize(context.Context, mcp.InitializeRequest) (*mcp.InitializeResult, error) {
	return new(mcp.InitializeResult), nil
}
func (*fakeMCPClient) ListTools(context.Context, mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	return new(mcp.ListToolsResult), nil
}
func (*fakeMCPClient) CallTool(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return new(mcp.CallToolResult), nil
}
func (*fakeMCPClient) Close() error { return nil }
