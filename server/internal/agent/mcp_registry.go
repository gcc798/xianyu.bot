package agent

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"gorm.io/gorm"
)

// connectedMCPClient 描述 MCP 发现和调用所需的客户端能力。
type connectedMCPClient interface {
	// Start 建立 MCP 传输。
	Start(context.Context) error
	// Initialize 协商 MCP 协议与客户端身份。
	Initialize(context.Context, mcp.InitializeRequest) (*mcp.InitializeResult, error)
	// ListTools 获取服务端工具目录。
	ListTools(context.Context, mcp.ListToolsRequest) (*mcp.ListToolsResult, error)
	// CallTool 调用远程工具。
	CallTool(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
	// Close 释放 MCP 传输资源。
	Close() error
}

// mcpConnection 保存已初始化服务、客户端和工具目录。
type mcpConnection struct {
	Server model.MCPServer    // Server 是持久化服务配置。
	Client connectedMCPClient // Client 是活动 MCP 客户端。
	Tools  []mcp.Tool         // Tools 是发现时返回的工具目录。
}

// Close 关闭底层 MCP 客户端。
func (c *mcpConnection) Close() error { return c.Client.Close() }

// MCPRegistry 发现所有启用的 MCP 服务并记录健康状态。
type MCPRegistry struct {
	db      *gorm.DB                                                       // db 读取服务配置并写入检查状态。
	connect func(context.Context, model.MCPServer) (*mcpConnection, error) // connect 是可测试替换的连接函数。
}

// NewMCPRegistry 创建 MCP 服务注册表。
func NewMCPRegistry(db *gorm.DB) *MCPRegistry {
	return &MCPRegistry{db: db}
}

// Discover 连接启用服务并返回成功连接；全部失败时合并错误。
func (r *MCPRegistry) Discover(ctx context.Context) ([]*mcpConnection, error) {
	servers := make([]model.MCPServer, 0)
	if err := r.db.WithContext(ctx).Where("enabled = ?", true).Order("id").Find(&servers).Error; err != nil {
		return nil, err
	}
	connect := r.connect
	if connect == nil {
		connect = connectMCPServer
	}
	connections := make([]*mcpConnection, 0, len(servers))
	failures := make([]error, 0)
	for _, server := range servers {
		connection, err := connect(ctx, server)
		if err != nil {
			failures = append(failures, fmt.Errorf("MCP %s: %w", server.Code, err))
			r.updateStatus(ctx, server.ID, "unavailable", err)
			continue
		}
		connections = append(connections, connection)
		r.updateStatus(ctx, server.ID, "available", nil)
	}
	if len(connections) == 0 && len(failures) > 0 {
		return nil, errors.Join(failures...)
	}
	return connections, nil
}

// updateStatus 保存服务最近检查时间、状态和截断后的错误。
func (r *MCPRegistry) updateStatus(ctx context.Context, serverID uint, status string, cause error) {
	now := time.Now()
	message := ""
	if cause != nil {
		message = cause.Error()
		if len(message) > 1000 {
			message = message[:1000]
		}
	}
	_ = r.db.WithContext(ctx).Model(&model.MCPServer{}).Where("id = ?", serverID).Updates(map[string]any{
		"last_status": status, "last_checked_at": now, "last_error": message,
	}).Error
}

// connectMCPServer 建立 HTTP 传输、初始化协议并获取工具目录。
func connectMCPServer(ctx context.Context, server model.MCPServer) (*mcpConnection, error) {
	headers := make(map[string]string)
	if header := strings.TrimSpace(server.AuthHeader); header != "" && server.AuthToken != "" {
		headers[header] = server.AuthToken
	}
	timeout := time.Duration(server.RequestTimeoutSeconds) * time.Second
	httpClient := &http.Client{Timeout: timeout}
	mcpClient, err := client.NewStreamableHttpClient(server.BaseURL,
		transport.WithHTTPBasicClient(httpClient), transport.WithHTTPHeaders(headers))
	if err != nil {
		return nil, err
	}
	closeOnError := func(err error) (*mcpConnection, error) {
		_ = mcpClient.Close()
		return nil, err
	}
	discoveryContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := mcpClient.Start(discoveryContext); err != nil {
		return closeOnError(err)
	}
	_, err = mcpClient.Initialize(discoveryContext, mcp.InitializeRequest{Params: mcp.InitializeParams{
		ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
		ClientInfo:      mcp.Implementation{Name: "xy-agent", Version: "1.0.0"},
	}})
	if err != nil {
		return closeOnError(err)
	}
	tools, err := mcpClient.ListTools(discoveryContext, mcp.ListToolsRequest{})
	if err != nil {
		return closeOnError(err)
	}
	return &mcpConnection{Server: server, Client: mcpClient, Tools: tools.Tools}, nil
}

// closeMCPConnections 关闭全部连接并合并关闭错误。
func closeMCPConnections(connections []*mcpConnection) error {
	errs := make([]error, 0)
	for _, connection := range connections {
		if err := connection.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
