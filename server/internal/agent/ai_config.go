package agent

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"gorm.io/gorm"
)

// mcpCodePattern 限制 MCP 服务编码为稳定的小写标识。
var mcpCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{1,63}$`)

// AIConfig 管理模型、MCP 服务和 Agent 全局运行限制。
type AIConfig struct {
	db          *gorm.DB     // db 持久化 Agent 配置。
	mcpRegistry *MCPRegistry // mcpRegistry 在配置变化后失效内存目录。
}

// NewAIConfig 创建 Agent 配置服务。
func NewAIConfig(db *gorm.DB, registries ...*MCPRegistry) *AIConfig {
	config := &AIConfig{db: db}
	if len(registries) > 0 {
		config.mcpRegistry = registries[0]
	}
	return config
}

// affected 将零行更新转换为记录不存在错误。
func affected(result *gorm.DB) error {
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListModels 按运行优先级列出模型，并标记凭据是否存在。
func (s *AIConfig) ListModels(ctx context.Context) ([]model.AIModel, error) {
	items := make([]model.AIModel, 0)
	err := s.db.WithContext(ctx).Order("priority ASC, id ASC").Find(&items).Error
	for i := range items {
		items[i].HasAPIKey = items[i].APIKey != ""
	}
	return items, err
}

// CreateModel 校验并创建模型配置。
func (s *AIConfig) CreateModel(ctx context.Context, input xyRequest.SaveAIModel) (model.AIModel, error) {
	item, err := buildAIModel(input)
	if err != nil {
		return model.AIModel{}, err
	}
	err = s.db.WithContext(ctx).Create(&item).Error
	item.HasAPIKey = item.APIKey != ""
	return item, err
}

// UpdateModel 更新模型配置，空 API Key 保留现有凭据。
func (s *AIConfig) UpdateModel(ctx context.Context, id uint, input xyRequest.SaveAIModel) error {
	if id == 0 {
		return errors.New("模型 ID 不能为空")
	}
	item, err := buildAIModel(input)
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current model.AIModel
		if err := tx.First(&current, id).Error; err != nil {
			return err
		}
		updates := map[string]any{
			"name": item.Name, "provider": item.Provider, "base_url": item.BaseURL, "model": item.Model,
			"enabled": item.Enabled, "priority": item.Priority, "supports_tools": item.SupportsTools,
			"supports_streaming": item.SupportsStreaming, "context_window": item.ContextWindow,
			"max_output_tokens": item.MaxOutputTokens, "request_timeout_seconds": item.RequestTimeoutSeconds,
			"extra_config": item.ExtraConfig,
		}
		if strings.TrimSpace(input.APIKey) != "" {
			updates["api_key"] = strings.TrimSpace(input.APIKey)
		}
		return tx.Model(&current).Updates(updates).Error
	})
}

// DeleteModel 删除指定模型配置。
func (s *AIConfig) DeleteModel(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("模型 ID 不能为空")
	}
	result := s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.AIModel{})
	return affected(result)
}

// ListMCPServers 列出 MCP 服务，并仅暴露认证令牌是否存在。
func (s *AIConfig) ListMCPServers(ctx context.Context) ([]model.MCPServer, error) {
	items := make([]model.MCPServer, 0)
	err := s.db.WithContext(ctx).Order("id DESC").Find(&items).Error
	for i := range items {
		items[i].HasAuthToken = items[i].AuthToken != ""
	}
	return items, err
}

// CreateMCPServer 校验并创建 MCP 服务配置。
func (s *AIConfig) CreateMCPServer(ctx context.Context, input xyRequest.SaveMCPServer) (model.MCPServer, error) {
	item, err := buildMCPServer(input)
	if err != nil {
		return model.MCPServer{}, err
	}
	err = s.db.WithContext(ctx).Create(&item).Error
	item.HasAuthToken = item.AuthToken != ""
	if err == nil {
		err = s.reloadMCPCatalog(ctx)
	}
	return item, err
}

// UpdateMCPServer 更新 MCP 服务，空认证令牌保留现有凭据。
func (s *AIConfig) UpdateMCPServer(ctx context.Context, id uint, input xyRequest.SaveMCPServer) error {
	if id == 0 {
		return errors.New("MCP 服务 ID 不能为空")
	}
	item, err := buildMCPServer(input)
	if err != nil {
		return err
	}
	updates := map[string]any{
		"name": item.Name, "code": item.Code, "description": item.Description, "tags": encodeMCPList(item.Tags),
		"base_url": item.BaseURL, "auth_header": item.AuthHeader,
		"enabled": item.Enabled, "request_timeout_seconds": item.RequestTimeoutSeconds,
		"task_timeout_seconds": item.TaskTimeoutSeconds, "enabled_tools": encodeMCPList(item.EnabledTools),
		"disabled_tools": encodeMCPList(item.DisabledTools), "catalog_ttl_seconds": item.CatalogTTLSeconds,
		"connection_ttl_seconds": item.ConnectionTTLSeconds, "connection_scope": item.ConnectionScope,
		"max_concurrent_calls": item.MaxConcurrentCalls, "catalog_status": "stale",
	}
	if strings.TrimSpace(input.AuthToken) != "" {
		updates["auth_token"] = strings.TrimSpace(input.AuthToken)
	}
	if err := affected(s.db.WithContext(ctx).Model(&model.MCPServer{}).Where("id = ?", id).Updates(updates)); err != nil {
		return err
	}
	return s.reloadMCPCatalog(ctx)
}

// DeleteMCPServer 删除指定 MCP 服务配置。
func (s *AIConfig) DeleteMCPServer(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("MCP 服务 ID 不能为空")
	}
	if err := affected(s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.MCPServer{})); err != nil {
		return err
	}
	return s.reloadMCPCatalog(ctx)
}

// CheckMCPServerHealth 执行指定服务的独立健康探测并返回最新投影。
func (s *AIConfig) CheckMCPServerHealth(ctx context.Context, id uint) (model.MCPServer, error) {
	server, err := s.mcpServer(ctx, id)
	if err != nil {
		return model.MCPServer{}, err
	}
	if s.mcpRegistry == nil {
		return model.MCPServer{}, errors.New("MCP 注册表未初始化")
	}
	checked, probeErr := s.mcpRegistry.CheckHealth(ctx, server.Code)
	if checked.ID == 0 {
		return model.MCPServer{}, probeErr
	}
	checked.HasAuthToken = checked.AuthToken != ""
	return checked, nil
}

// RefreshMCPServerCatalog 刷新指定服务目录并返回成功或失败后的最新投影。
func (s *AIConfig) RefreshMCPServerCatalog(ctx context.Context, id uint) (model.MCPServer, error) {
	server, err := s.mcpServer(ctx, id)
	if err != nil {
		return model.MCPServer{}, err
	}
	if s.mcpRegistry == nil {
		return model.MCPServer{}, errors.New("MCP 注册表未初始化")
	}
	refreshErr := s.mcpRegistry.RefreshCatalog(ctx, server.Code)
	latest, err := s.mcpServer(ctx, id)
	if err != nil {
		return model.MCPServer{}, errors.Join(refreshErr, err)
	}
	latest.HasAuthToken = latest.AuthToken != ""
	return latest, nil
}

// ListMCPServerTools 返回指定服务当前获准的缓存工具投影。
func (s *AIConfig) ListMCPServerTools(ctx context.Context, id uint) ([]model.MCPTool, error) {
	server, err := s.mcpServer(ctx, id)
	if err != nil {
		return nil, err
	}
	if s.mcpRegistry == nil {
		return nil, errors.New("MCP 注册表未初始化")
	}
	return s.mcpRegistry.CachedTools(server.Code), nil
}

// mcpServer 按主键读取 MCP 服务配置。
func (s *AIConfig) mcpServer(ctx context.Context, id uint) (model.MCPServer, error) {
	if id == 0 {
		return model.MCPServer{}, errors.New("MCP 服务 ID 不能为空")
	}
	var server model.MCPServer
	err := s.db.WithContext(ctx).First(&server, id).Error
	return server, err
}

// reloadMCPCatalog 让配置变更立即作用于路由过滤。
func (s *AIConfig) reloadMCPCatalog(ctx context.Context) error {
	if s.mcpRegistry == nil {
		return nil
	}
	return s.mcpRegistry.LoadCatalog(ctx)
}

// GetAgentConfig 返回唯一运行限制，不存在时创建默认值。
func (s *AIConfig) GetAgentConfig(ctx context.Context) (model.AgentConfig, error) {
	item := model.DefaultAgentConfig()
	err := s.db.WithContext(ctx).FirstOrCreate(&item, model.AgentConfig{ID: 1}).Error
	return item, err
}

// UpdateAgentConfig 校验并覆盖唯一 Agent 运行限制。
func (s *AIConfig) UpdateAgentConfig(ctx context.Context, item model.AgentConfig) error {
	item.ID = 1
	if err := validateAgentConfig(item); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Save(&item).Error
}

// buildAIModel 规范化并校验模型地址、能力、令牌限制和扩展 JSON。
func buildAIModel(input xyRequest.SaveAIModel) (model.AIModel, error) {
	item := model.AIModel{
		Name: strings.TrimSpace(input.Name), Provider: strings.TrimSpace(input.Provider), BaseURL: strings.TrimSpace(input.BaseURL),
		APIKey: strings.TrimSpace(input.APIKey), Model: strings.TrimSpace(input.Model), Enabled: input.Enabled,
		Priority: input.Priority, SupportsTools: input.SupportsTools, SupportsStreaming: input.SupportsStreaming,
		ContextWindow: input.ContextWindow, MaxOutputTokens: input.MaxOutputTokens,
		RequestTimeoutSeconds: input.RequestTimeoutSeconds, ExtraConfig: strings.TrimSpace(input.ExtraConfig),
	}
	if item.Name == "" || item.Provider == "" || item.Model == "" {
		return model.AIModel{}, errors.New("模型名称、提供商和模型标识不能为空")
	}
	if err := validateHTTPURL(item.BaseURL); err != nil {
		return model.AIModel{}, err
	}
	if item.Priority < 1 {
		return model.AIModel{}, errors.New("模型优先级必须为正整数")
	}
	if item.ContextWindow < 1 || item.MaxOutputTokens < 1 || item.RequestTimeoutSeconds < 1 || item.RequestTimeoutSeconds > 600 {
		return model.AIModel{}, errors.New("模型上下文、输出 Token 和超时时间必须为有效正数")
	}
	if item.ExtraConfig != "" && !json.Valid([]byte(item.ExtraConfig)) {
		return model.AIModel{}, errors.New("扩展配置必须是有效 JSON")
	}
	return item, nil
}

// buildMCPServer 规范化并校验 MCP 编码、地址和超时。
func buildMCPServer(input xyRequest.SaveMCPServer) (model.MCPServer, error) {
	item := model.MCPServer{
		Name: strings.TrimSpace(input.Name), Code: strings.TrimSpace(input.Code), Description: strings.TrimSpace(input.Description),
		Tags: normalizeMCPList(input.Tags), BaseURL: strings.TrimSpace(input.BaseURL),
		AuthHeader: strings.TrimSpace(input.AuthHeader), AuthToken: strings.TrimSpace(input.AuthToken), Enabled: input.Enabled,
		EnabledTools: normalizeMCPList(input.EnabledTools), DisabledTools: normalizeMCPList(input.DisabledTools),
		RequestTimeoutSeconds: input.RequestTimeoutSeconds, TaskTimeoutSeconds: input.TaskTimeoutSeconds,
		CatalogTTLSeconds: input.CatalogTTLSeconds, ConnectionTTLSeconds: input.ConnectionTTLSeconds,
		ConnectionScope: strings.TrimSpace(input.ConnectionScope), MaxConcurrentCalls: input.MaxConcurrentCalls,
		CatalogStatus: "unknown", LastStatus: "unknown",
	}
	if item.Name == "" || !mcpCodePattern.MatchString(item.Code) {
		return model.MCPServer{}, errors.New("服务名称不能为空，编码必须以小写字母开头且只包含小写字母、数字、下划线或连字符")
	}
	if err := validateHTTPURL(item.BaseURL); err != nil {
		return model.MCPServer{}, err
	}
	if item.RequestTimeoutSeconds < 1 || item.RequestTimeoutSeconds > 600 || item.TaskTimeoutSeconds < 1 || item.TaskTimeoutSeconds > 86400 {
		return model.MCPServer{}, errors.New("MCP 请求或任务超时时间无效")
	}
	if len(item.Description) > 1000 || len(item.Tags) > 32 || len(item.EnabledTools) > 500 || len(item.DisabledTools) > 500 {
		return model.MCPServer{}, errors.New("MCP 描述、标签或工具筛选配置超出允许范围")
	}
	if item.CatalogTTLSeconds < 10 || item.CatalogTTLSeconds > 86400 || item.ConnectionTTLSeconds < 10 || item.ConnectionTTLSeconds > 86400 {
		return model.MCPServer{}, errors.New("MCP 目录或连接 TTL 必须在 10 到 86400 秒之间")
	}
	if item.ConnectionScope != "shared" && item.ConnectionScope != "owner" {
		return model.MCPServer{}, errors.New("MCP 连接作用域必须是 shared 或 owner")
	}
	if item.MaxConcurrentCalls < 0 || item.MaxConcurrentCalls > 1000 {
		return model.MCPServer{}, errors.New("MCP 调用并发上限必须在 0 到 1000 之间")
	}
	for _, tool := range item.EnabledTools {
		if slices.Contains(item.DisabledTools, tool) {
			return model.MCPServer{}, errors.New("同一 MCP 工具不能同时出现在允许和禁止列表")
		}
	}
	return item, nil
}

// normalizeMCPList 清理、去重并稳定保存 MCP 标签或工具名称。
func normalizeMCPList(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && len(value) <= 120 && !slices.Contains(result, value) {
			result = append(result, value)
		}
	}
	return result
}

// encodeMCPList 为 map 更新路径生成稳定的 JSON 列表。
func encodeMCPList(values []string) string {
	data, _ := json.Marshal(values)
	return string(data)
}

// validateHTTPURL 仅接受带主机的 HTTP 或 HTTPS URL。
func validateHTTPURL(value string) error {
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("服务地址必须是有效的 HTTP 或 HTTPS URL")
	}
	return nil
}

// validateAgentConfig 校验运行限制为正数且不超过服务安全上限。
func validateAgentConfig(item model.AgentConfig) error {
	values := []int{item.MaxWallTimeSeconds, item.MaxIterations, item.MaxModelCalls, item.MaxToolCalls, item.MaxConcurrentToolCalls, item.MaxSearchCalls, item.MaxWriteToolCalls, item.MaxOutputTokens, item.ConversationHistoryLimit, item.SummaryTriggerTokens}
	for _, value := range values {
		if value < 1 {
			return errors.New("Agent 配置项必须为正整数")
		}
	}
	if item.MaxWallTimeSeconds > 600 || item.MaxIterations > 50 || item.MaxModelCalls > 100 || item.MaxToolCalls > 200 {
		return errors.New("Agent 运行限制超出允许范围")
	}
	if item.MaxConcurrentToolCalls > 20 || item.MaxConcurrentToolCalls > item.MaxToolCalls {
		return errors.New("Agent 工具并发数超出允许范围")
	}
	if item.MaxSearchCalls > item.MaxToolCalls || item.MaxWriteToolCalls > item.MaxToolCalls {
		return errors.New("搜索和写工具调用次数不能超过工具调用总次数")
	}
	return nil
}
