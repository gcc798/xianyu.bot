package agent

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
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

// mcpHealthClient 描述独立健康探测所需的最小协议能力。
type mcpHealthClient interface {
	Start(context.Context) error
	Initialize(context.Context, mcp.InitializeRequest) (*mcp.InitializeResult, error)
	Ping(context.Context) error
	Close() error
}

// mcpConnection 保存已初始化服务、客户端和工具目录。
type mcpConnection struct {
	Server       model.MCPServer    // Server 是持久化服务配置。
	Client       connectedMCPClient // Client 是活动 MCP 客户端。
	Tools        []mcp.Tool         // Tools 是发现时返回的完整分页工具目录。
	Instructions string             // Instructions 是 MCP 握手披露的不可信说明。
}

// Close 关闭底层 MCP 客户端。
func (c *mcpConnection) Close() error { return c.Client.Close() }

// MCPRegistry 发现所有启用的 MCP 服务并记录健康状态。
type MCPRegistry struct {
	db                *gorm.DB                                                           // db 读取服务配置并写入检查状态。
	connect           func(context.Context, model.MCPServer) (*mcpConnection, error)     // connect 是可测试替换的连接函数。
	healthClient      func(model.MCPServer) (mcpHealthClient, error)                     // healthClient 创建独立短期探测客户端。
	catalog           atomic.Pointer[mcpCatalogSnapshot]                                 // catalog 保存原子发布的只读工具目录。
	catalogRefreshMu  sync.Mutex                                                         // catalogRefreshMu 合并重叠的目录扫描。
	healthMu          sync.Mutex                                                         // healthMu 保护同服务健康探测合并状态。
	healthChecks      map[string]*mcpHealthCheck                                         // healthChecks 保存正在执行的健康探测。
	runtimeConnect    func(context.Context, model.MCPServer) (connectedMCPClient, error) // runtimeConnect 创建惰性执行客户端。
	runtimeMu         sync.Mutex                                                         // runtimeMu 保护执行连接表和建连状态。
	runtimeClients    map[string]*mcpRuntimeConnection                                   // runtimeClients 保存各连接作用域的活动客户端。
	runtimeConnecting map[string]*mcpRuntimeConnect                                      // runtimeConnecting 合并同作用域首次建连。
	runtimeGeneration atomic.Uint64                                                      // runtimeGeneration 生成执行连接代次。
	runtimeClosing    bool                                                               // runtimeClosing 阻止服务退出后建立新连接。
}

// NewMCPRegistry 创建 MCP 服务注册表。
func NewMCPRegistry(db *gorm.DB) *MCPRegistry {
	return &MCPRegistry{
		db: db, healthChecks: make(map[string]*mcpHealthCheck),
		runtimeClients: make(map[string]*mcpRuntimeConnection), runtimeConnecting: make(map[string]*mcpRuntimeConnect),
	}
}

// connectMCPServer 建立 HTTP 传输、初始化协议并获取工具目录。
func connectMCPServer(ctx context.Context, server model.MCPServer) (*mcpConnection, error) {
	mcpClient, err := newMCPHTTPClient(server)
	if err != nil {
		return nil, err
	}
	timeout := time.Duration(server.RequestTimeoutSeconds) * time.Second
	closeOnError := func(err error) (*mcpConnection, error) {
		_ = mcpClient.Close()
		return nil, err
	}
	discoveryContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := mcpClient.Start(discoveryContext); err != nil {
		return closeOnError(err)
	}
	initialized, err := mcpClient.Initialize(discoveryContext, mcp.InitializeRequest{Params: mcp.InitializeParams{
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
	return &mcpConnection{Server: server, Client: mcpClient, Tools: tools.Tools, Instructions: initialized.Instructions}, nil
}

// newMCPHTTPClient 创建带服务超时和认证头的 Streamable HTTP 客户端。
func newMCPHTTPClient(server model.MCPServer) (*client.Client, error) {
	headers := make(map[string]string)
	if header := strings.TrimSpace(server.AuthHeader); header != "" && server.AuthToken != "" {
		headers[header] = server.AuthToken
	}
	timeout := time.Duration(server.RequestTimeoutSeconds) * time.Second
	httpClient := &http.Client{Timeout: timeout}
	return client.NewStreamableHttpClient(server.BaseURL,
		transport.WithHTTPBasicClient(httpClient), transport.WithHTTPHeaders(headers))
}
