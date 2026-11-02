package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"gorm.io/gorm"
)

var mcpCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{1,63}$`)

type aiConfig struct {
	db *gorm.DB
}

func (s *aiConfig) ListModels(ctx context.Context) ([]model.AIModel, error) {
	items := make([]model.AIModel, 0)
	err := s.db.WithContext(ctx).Order("is_default DESC, id DESC").Find(&items).Error
	for i := range items {
		items[i].HasAPIKey = items[i].APIKey != ""
	}
	return items, err
}

func (s *aiConfig) CreateModel(ctx context.Context, input xyRequest.SaveAIModel) (model.AIModel, error) {
	item, err := buildAIModel(input)
	if err != nil {
		return model.AIModel{}, err
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if item.IsDefault {
			if err := tx.Model(&model.AIModel{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Create(&item).Error
	})
	item.HasAPIKey = item.APIKey != ""
	return item, err
}

func (s *aiConfig) UpdateModel(ctx context.Context, id uint, input xyRequest.SaveAIModel) error {
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
		if item.IsDefault {
			if err := tx.Model(&model.AIModel{}).Where("id <> ?", id).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		updates := map[string]any{
			"name": item.Name, "provider": item.Provider, "base_url": item.BaseURL, "model": item.Model,
			"enabled": item.Enabled, "is_default": item.IsDefault, "supports_tools": item.SupportsTools,
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

func (s *aiConfig) DeleteModel(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("模型 ID 不能为空")
	}
	result := s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.AIModel{})
	return affected(result)
}

func (s *aiConfig) SetDefaultModel(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("模型 ID 不能为空")
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item model.AIModel
		if err := tx.First(&item, id).Error; err != nil {
			return err
		}
		if !item.Enabled || !item.SupportsTools {
			return errors.New("默认模型必须已启用并支持工具调用")
		}
		if err := tx.Model(&model.AIModel{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return err
		}
		return tx.Model(&item).Update("is_default", true).Error
	})
}

func (s *aiConfig) ListMCPServers(ctx context.Context) ([]model.MCPServer, error) {
	items := make([]model.MCPServer, 0)
	err := s.db.WithContext(ctx).Order("id DESC").Find(&items).Error
	for i := range items {
		items[i].HasAuthToken = items[i].AuthToken != ""
	}
	return items, err
}

func (s *aiConfig) CreateMCPServer(ctx context.Context, input xyRequest.SaveMCPServer) (model.MCPServer, error) {
	item, err := buildMCPServer(input)
	if err != nil {
		return model.MCPServer{}, err
	}
	err = s.db.WithContext(ctx).Create(&item).Error
	item.HasAuthToken = item.AuthToken != ""
	return item, err
}

func (s *aiConfig) UpdateMCPServer(ctx context.Context, id uint, input xyRequest.SaveMCPServer) error {
	if id == 0 {
		return errors.New("MCP 服务 ID 不能为空")
	}
	item, err := buildMCPServer(input)
	if err != nil {
		return err
	}
	updates := map[string]any{
		"name": item.Name, "code": item.Code, "base_url": item.BaseURL, "auth_header": item.AuthHeader,
		"enabled": item.Enabled, "request_timeout_seconds": item.RequestTimeoutSeconds,
		"task_timeout_seconds": item.TaskTimeoutSeconds,
	}
	if strings.TrimSpace(input.AuthToken) != "" {
		updates["auth_token"] = strings.TrimSpace(input.AuthToken)
	}
	return affected(s.db.WithContext(ctx).Model(&model.MCPServer{}).Where("id = ?", id).Updates(updates))
}

func (s *aiConfig) DeleteMCPServer(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("MCP 服务 ID 不能为空")
	}
	return affected(s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.MCPServer{}))
}

func (s *aiConfig) GetAgentConfig(ctx context.Context) (model.AgentConfig, error) {
	item := model.DefaultAgentConfig()
	err := s.db.WithContext(ctx).FirstOrCreate(&item, model.AgentConfig{ID: 1}).Error
	return item, err
}

func (s *aiConfig) UpdateAgentConfig(ctx context.Context, item model.AgentConfig) error {
	item.ID = 1
	if err := validateAgentConfig(item); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Save(&item).Error
}

func buildAIModel(input xyRequest.SaveAIModel) (model.AIModel, error) {
	item := model.AIModel{
		Name: strings.TrimSpace(input.Name), Provider: strings.TrimSpace(input.Provider), BaseURL: strings.TrimSpace(input.BaseURL),
		APIKey: strings.TrimSpace(input.APIKey), Model: strings.TrimSpace(input.Model), Enabled: input.Enabled,
		IsDefault: input.IsDefault, SupportsTools: input.SupportsTools, SupportsStreaming: input.SupportsStreaming,
		ContextWindow: input.ContextWindow, MaxOutputTokens: input.MaxOutputTokens,
		RequestTimeoutSeconds: input.RequestTimeoutSeconds, ExtraConfig: strings.TrimSpace(input.ExtraConfig),
	}
	if item.Name == "" || item.Provider == "" || item.Model == "" {
		return model.AIModel{}, errors.New("模型名称、提供商和模型标识不能为空")
	}
	if err := validateHTTPURL(item.BaseURL); err != nil {
		return model.AIModel{}, err
	}
	if item.ContextWindow < 1 || item.MaxOutputTokens < 1 || item.RequestTimeoutSeconds < 1 || item.RequestTimeoutSeconds > 600 {
		return model.AIModel{}, errors.New("模型上下文、输出 Token 和超时时间必须为有效正数")
	}
	if item.ExtraConfig != "" && !json.Valid([]byte(item.ExtraConfig)) {
		return model.AIModel{}, errors.New("扩展配置必须是有效 JSON")
	}
	if item.IsDefault && (!item.Enabled || !item.SupportsTools) {
		return model.AIModel{}, errors.New("默认模型必须已启用并支持工具调用")
	}
	return item, nil
}

func buildMCPServer(input xyRequest.SaveMCPServer) (model.MCPServer, error) {
	item := model.MCPServer{
		Name: strings.TrimSpace(input.Name), Code: strings.TrimSpace(input.Code), BaseURL: strings.TrimSpace(input.BaseURL),
		AuthHeader: strings.TrimSpace(input.AuthHeader), AuthToken: strings.TrimSpace(input.AuthToken), Enabled: input.Enabled,
		RequestTimeoutSeconds: input.RequestTimeoutSeconds, TaskTimeoutSeconds: input.TaskTimeoutSeconds, LastStatus: "unknown",
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
	return item, nil
}

func validateHTTPURL(value string) error {
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("服务地址必须是有效的 HTTP 或 HTTPS URL")
	}
	return nil
}

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
