package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/mark3labs/mcp-go/mcp"
	"gorm.io/gorm"
)

const (
	maxMCPTools          = 1000
	maxMCPDefinitionSize = 512 << 10
	maxMCPDescription    = 4000
)

var mcpToolNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.:-]{0,159}$`)

// RefreshCatalog 使用短期连接刷新指定 MCP 的完整工具目录。
func (r *MCPRegistry) RefreshCatalog(ctx context.Context, serverCode string) error {
	var server model.MCPServer
	if err := r.db.WithContext(ctx).Where("code = ? AND enabled = ?", serverCode, true).First(&server).Error; err != nil {
		return err
	}
	connect := r.connect
	if connect == nil {
		connect = connectMCPServer
	}
	connection, err := connect(ctx, server)
	if err != nil {
		r.recordCatalogFailure(ctx, server, err)
		return err
	}
	defer connection.Close()

	discoveredAt := time.Now()
	tools, warnings, err := catalogModels(server.Code, connection.Tools, discoveredAt)
	if err != nil {
		r.recordCatalogFailure(ctx, server, err)
		return err
	}
	warning := strings.Join(warnings, "; ")
	warning = truncateMCPText(warning, 1000)
	instructions := truncateMCPText(connection.Instructions, maxMCPDescription)
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("mcp_server_code = ?", server.Code).Delete(&model.MCPTool{}).Error; err != nil {
			return err
		}
		if len(tools) > 0 {
			if err := tx.Create(&tools).Error; err != nil {
				return err
			}
		}
		return tx.Model(&model.MCPServer{}).Where("id = ?", server.ID).Updates(map[string]any{
			"server_instructions": instructions, "catalog_status": "available", "catalog_tool_count": len(tools),
			"catalog_refreshed_at": discoveredAt, "catalog_error": warning,
		}).Error
	})
	if err != nil {
		r.recordCatalogFailure(ctx, server, err)
		return err
	}
	return r.LoadCatalog(ctx)
}

// catalogModels 校验外部工具并转换为可持久化目录记录。
func catalogModels(serverCode string, remoteTools []mcp.Tool, discoveredAt time.Time) ([]model.MCPTool, []string, error) {
	if len(remoteTools) > maxMCPTools {
		return nil, nil, fmt.Errorf("MCP 工具数量 %d 超过上限 %d", len(remoteTools), maxMCPTools)
	}
	tools := make([]model.MCPTool, 0, len(remoteTools))
	warnings := make([]string, 0)
	seen := make(map[string]struct{}, len(remoteTools))
	for _, remote := range remoteTools {
		tool, err := catalogModel(serverCode, remote, discoveredAt)
		if err != nil {
			warnings = append(warnings, truncateMCPText(err.Error(), 200))
			continue
		}
		if _, duplicate := seen[tool.Name]; duplicate {
			warnings = append(warnings, "跳过重复工具 "+tool.Name)
			continue
		}
		seen[tool.Name] = struct{}{}
		tools = append(tools, tool)
	}
	if len(remoteTools) > 0 && len(tools) == 0 {
		return nil, warnings, errors.New("MCP 未返回可安全缓存的工具定义")
	}
	return tools, warnings, nil
}

// catalogModel 构造一个经过大小、名称和 JSON 校验的目录记录。
func catalogModel(serverCode string, remote mcp.Tool, discoveredAt time.Time) (model.MCPTool, error) {
	if !mcpToolNamePattern.MatchString(remote.Name) {
		return model.MCPTool{}, fmt.Errorf("跳过无效工具名称 %q", truncateMCPText(remote.Name, 80))
	}
	definition, err := json.Marshal(remote)
	if err != nil || !json.Valid(definition) || len(definition) > maxMCPDefinitionSize {
		return model.MCPTool{}, fmt.Errorf("跳过工具 %s：定义无效或过大", remote.Name)
	}
	inputSchema, err := mcpInputSchema(remote)
	if err != nil {
		return model.MCPTool{}, fmt.Errorf("跳过工具 %s：输入 Schema 无效", remote.Name)
	}
	outputSchema, err := mcpOutputSchema(remote)
	if err != nil {
		return model.MCPTool{}, fmt.Errorf("跳过工具 %s：输出 Schema 无效", remote.Name)
	}
	annotations, err := json.Marshal(remote.Annotations)
	if err != nil {
		return model.MCPTool{}, fmt.Errorf("跳过工具 %s：annotations 无效", remote.Name)
	}
	hash := sha256.Sum256(definition)
	description := truncateMCPText(remote.Description, maxMCPDescription)
	return model.MCPTool{
		MCPServerCode: serverCode, Name: remote.Name, Description: description,
		InputSchema: string(inputSchema), OutputSchema: string(outputSchema), Annotations: string(annotations),
		Definition: string(definition), SearchText: catalogSearchText(remote.Name, description),
		DefinitionHash: hex.EncodeToString(hash[:]), DiscoveredAt: discoveredAt,
	}, nil
}

// mcpInputSchema 返回工具实际使用的输入 Schema JSON。
func mcpInputSchema(tool mcp.Tool) ([]byte, error) {
	if len(tool.RawInputSchema) > 0 {
		if !json.Valid(tool.RawInputSchema) || len(tool.RawInputSchema) > maxMCPDefinitionSize {
			return nil, errors.New("输入 Schema 无效")
		}
		return slices.Clone(tool.RawInputSchema), nil
	}
	return json.Marshal(tool.InputSchema)
}

// mcpOutputSchema 返回工具实际使用的输出 Schema JSON。
func mcpOutputSchema(tool mcp.Tool) ([]byte, error) {
	if len(tool.RawOutputSchema) > 0 {
		if !json.Valid(tool.RawOutputSchema) || len(tool.RawOutputSchema) > maxMCPDefinitionSize {
			return nil, errors.New("输出 Schema 无效")
		}
		return slices.Clone(tool.RawOutputSchema), nil
	}
	if tool.OutputSchema.Type == "" {
		return []byte("{}"), nil
	}
	return json.Marshal(tool.OutputSchema)
}

// catalogSearchText 生成不包含 Schema 和服务指令的紧凑检索文本。
func catalogSearchText(name, description string) string {
	return truncateMCPText(strings.Join(strings.Fields(name+" "+description), " "), maxMCPDescription)
}

// recordCatalogFailure 保留旧目录，仅更新安全错误投影。
func (r *MCPRegistry) recordCatalogFailure(ctx context.Context, server model.MCPServer, cause error) {
	_ = r.db.WithContext(ctx).Model(&model.MCPServer{}).Where("id = ?", server.ID).Updates(map[string]any{
		"catalog_status": "unavailable", "catalog_error": safeMCPError(server, cause),
	}).Error
}

// safeMCPError 去除 MCP 凭据和 URL 用户信息后截断错误摘要。
func safeMCPError(server model.MCPServer, cause error) string {
	if cause == nil {
		return ""
	}
	message := cause.Error()
	if server.AuthToken != "" {
		message = strings.ReplaceAll(message, server.AuthToken, "[redacted]")
	}
	if parsed, err := url.Parse(server.BaseURL); err == nil && parsed.Host != "" {
		safeURL := *parsed
		safeURL.User = nil
		safeURL.RawQuery = ""
		safeURL.Fragment = ""
		message = strings.ReplaceAll(message, server.BaseURL, safeURL.String())
	}
	return truncateMCPText(message, 1000)
}

// truncateMCPText 按字符截断外部文本，避免切断 UTF-8。
func truncateMCPText(value string, limit int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) > limit {
		return string(runes[:limit])
	}
	return value
}
