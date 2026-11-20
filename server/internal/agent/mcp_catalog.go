package agent

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/mark3labs/mcp-go/mcp"
)

// mcpCatalogTool 是内存目录中的一个已校验工具。
type mcpCatalogTool struct {
	Server   model.MCPServer // Server 是工具所属的服务配置快照。
	Snapshot model.MCPTool   // Snapshot 是持久化工具元数据。
	Remote   mcp.Tool        // Remote 是可组装为 Agent Tool 的完整定义。
}

// Code 返回 Agent 使用的稳定工具编码。
func (t mcpCatalogTool) Code() string { return t.Server.Code + "__" + t.Remote.Name }

// mcpCatalogSnapshot 保存一次原子发布的只读 MCP 目录。
type mcpCatalogSnapshot struct {
	tools  []mcpCatalogTool
	byCode map[string]mcpCatalogTool
}

// LoadCatalog 从数据库快照重建并原子发布 MCP 内存目录。
func (r *MCPRegistry) LoadCatalog(ctx context.Context) error {
	servers := make([]model.MCPServer, 0)
	if err := r.db.WithContext(ctx).Where("enabled = ?", true).Order("id").Find(&servers).Error; err != nil {
		return err
	}
	tools := make([]model.MCPTool, 0)
	if err := r.db.WithContext(ctx).Order("id").Find(&tools).Error; err != nil {
		return err
	}
	r.catalog.Store(buildMCPCatalog(servers, tools))
	return nil
}

// CatalogTools 返回当前内存目录的顶层副本，查询不会连接 MCP。
func (r *MCPRegistry) CatalogTools() []mcpCatalogTool {
	snapshot := r.catalog.Load()
	if snapshot == nil {
		return nil
	}
	result := make([]mcpCatalogTool, len(snapshot.tools))
	for i, tool := range snapshot.tools {
		result[i] = cloneCatalogTool(tool)
	}
	return result
}

// CatalogTool 按 Agent 工具编码读取当前目录定义。
func (r *MCPRegistry) CatalogTool(code string) (mcpCatalogTool, bool) {
	snapshot := r.catalog.Load()
	if snapshot == nil {
		return mcpCatalogTool{}, false
	}
	tool, ok := snapshot.byCode[code]
	return cloneCatalogTool(tool), ok
}

// cloneCatalogTool 防止调用方修改原子快照中的切片或 Schema map。
func cloneCatalogTool(tool mcpCatalogTool) mcpCatalogTool {
	tool.Server.Tags = slices.Clone(tool.Server.Tags)
	tool.Server.EnabledTools = slices.Clone(tool.Server.EnabledTools)
	tool.Server.DisabledTools = slices.Clone(tool.Server.DisabledTools)
	tool.Remote = mcp.Tool{}
	_ = json.Unmarshal([]byte(tool.Snapshot.Definition), &tool.Remote)
	return tool
}

// CachedTools 返回指定 MCP 的安全缓存工具投影。
func (r *MCPRegistry) CachedTools(serverCode string) []model.MCPTool {
	result := make([]model.MCPTool, 0)
	for _, tool := range r.CatalogTools() {
		if tool.Server.Code == serverCode {
			result = append(result, tool.Snapshot)
		}
	}
	return result
}

// buildMCPCatalog 校验数据库记录并生成不可变目录；坏记录和重复编码只隔离自身。
func buildMCPCatalog(servers []model.MCPServer, tools []model.MCPTool) *mcpCatalogSnapshot {
	serverByCode := make(map[string]model.MCPServer, len(servers))
	for _, server := range servers {
		if server.Enabled {
			serverByCode[server.Code] = server
		}
	}
	snapshot := &mcpCatalogSnapshot{tools: make([]mcpCatalogTool, 0, len(tools)), byCode: make(map[string]mcpCatalogTool, len(tools))}
	for _, stored := range tools {
		server, ok := serverByCode[stored.MCPServerCode]
		if !ok || !mcpToolAllowed(server, stored.Name) {
			continue
		}
		var remote mcp.Tool
		if stored.Name == "" || !json.Valid([]byte(stored.Definition)) || json.Unmarshal([]byte(stored.Definition), &remote) != nil || remote.Name != stored.Name {
			continue
		}
		tool := mcpCatalogTool{Server: server, Snapshot: stored, Remote: remote}
		if _, duplicate := snapshot.byCode[tool.Code()]; duplicate {
			continue
		}
		snapshot.tools = append(snapshot.tools, tool)
		snapshot.byCode[tool.Code()] = tool
	}
	return snapshot
}

// mcpToolAllowed 在模型路由前应用服务 allow/deny 策略。
func mcpToolAllowed(server model.MCPServer, name string) bool {
	if !server.Enabled || slices.Contains(server.DisabledTools, name) {
		return false
	}
	return len(server.EnabledTools) == 0 || slices.Contains(server.EnabledTools, name)
}
