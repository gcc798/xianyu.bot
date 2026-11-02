package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gcc798/go.xianyu.assistant/server/internal/infra"
	"github.com/gcc798/go.xianyu.assistant/server/internal/model/xianyu"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

type connectedMCPClient interface {
	Start(context.Context) error
	Initialize(context.Context, mcp.InitializeRequest) (*mcp.InitializeResult, error)
	ListTools(context.Context, mcp.ListToolsRequest) (*mcp.ListToolsResult, error)
	CallTool(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
	Close() error
}

type mcpConnection struct {
	Server model.MCPServer
	Client connectedMCPClient
	Tools  []mcp.Tool
}

func (c *mcpConnection) Close() error { return c.Client.Close() }

type mcpRegistry struct {
	connect func(context.Context, model.MCPServer) (*mcpConnection, error)
}

func (r *mcpRegistry) Discover(ctx context.Context) ([]*mcpConnection, error) {
	servers := make([]model.MCPServer, 0)
	if err := infra.DB.WithContext(ctx).Where("enabled = ?", true).Order("id").Find(&servers).Error; err != nil {
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

func (r *mcpRegistry) updateStatus(ctx context.Context, serverID uint, status string, cause error) {
	now := time.Now()
	message := ""
	if cause != nil {
		message = cause.Error()
		if len(message) > 1000 {
			message = message[:1000]
		}
	}
	_ = infra.DB.WithContext(ctx).Model(&model.MCPServer{}).Where("id = ?", serverID).Updates(map[string]any{
		"last_status": status, "last_checked_at": now, "last_error": message,
	}).Error
}

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

func closeMCPConnections(connections []*mcpConnection) error {
	errs := make([]error, 0)
	for _, connection := range connections {
		if err := connection.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
