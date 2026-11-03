package model

import (
	"time"
)

// AIModel 保存一个可供 Agent 使用的大模型配置。
type AIModel struct {
	// Timestamps 保存模型配置的生命周期字段。
	Timestamps
	// Name 是配置的显示名称。
	Name string `json:"name" gorm:"size:80;not null"`
	// Provider 标识模型服务商或兼容协议。
	Provider string `json:"provider" gorm:"size:40;not null"`
	// BaseURL 是模型 API 的基础地址。
	BaseURL string `json:"baseUrl" gorm:"size:500;not null"`
	// APIKey 是模型服务凭据，不向客户端返回。
	APIKey string `json:"-" gorm:"type:text"`
	// Model 是服务商侧模型标识。
	Model string `json:"model" gorm:"size:120;not null"`
	// Enabled 决定配置是否可被选择。
	Enabled bool `json:"enabled" gorm:"not null;default:true"`
	// IsDefault 标识未指定模型时使用的配置。
	IsDefault bool `json:"isDefault" gorm:"not null;default:false;index"`
	// SupportsTools 表示模型是否支持工具调用。
	SupportsTools bool `json:"supportsTools" gorm:"not null;default:true"`
	// SupportsStreaming 表示模型是否支持流式输出。
	SupportsStreaming bool `json:"supportsStreaming" gorm:"not null;default:true"`
	// ContextWindow 是模型最大上下文令牌数。
	ContextWindow int `json:"contextWindow" gorm:"not null;default:128000"`
	// MaxOutputTokens 是单次生成的最大令牌数。
	MaxOutputTokens int `json:"maxOutputTokens" gorm:"not null;default:2000"`
	// RequestTimeoutSeconds 是单次模型请求超时秒数。
	RequestTimeoutSeconds int `json:"requestTimeoutSeconds" gorm:"not null;default:60"`
	// ExtraConfig 保存服务商专属的 JSON 配置。
	ExtraConfig string `json:"extraConfig" gorm:"type:text"`
	// HasAPIKey 表示配置是否已有凭据，仅用于响应投影。
	HasAPIKey bool `json:"hasApiKey" gorm:"-"`
}

// TableName 返回模型配置表名。
func (AIModel) TableName() string { return "xy_ai_models" }

// MCPServer 保存一个 Agent 可访问的 MCP 服务配置。
type MCPServer struct {
	// Timestamps 保存服务配置的生命周期字段。
	Timestamps
	// Name 是 MCP 服务显示名称。
	Name string `json:"name" gorm:"size:80;not null"`
	// Code 是 MCP 服务的稳定唯一编码。
	Code string `json:"code" gorm:"size:64;not null;uniqueIndex"`
	// BaseURL 是 MCP HTTP 入口地址。
	BaseURL string `json:"baseUrl" gorm:"size:500;not null"`
	// AuthHeader 是认证令牌使用的请求头名称。
	AuthHeader string `json:"authHeader" gorm:"size:100"`
	// AuthToken 是 MCP 认证凭据，不向客户端返回。
	AuthToken string `json:"-" gorm:"type:text"`
	// Enabled 决定服务及其工具是否可用。
	Enabled bool `json:"enabled" gorm:"not null;default:true"`
	// RequestTimeoutSeconds 是普通 MCP 请求超时秒数。
	RequestTimeoutSeconds int `json:"requestTimeoutSeconds" gorm:"not null;default:15"`
	// TaskTimeoutSeconds 是长任务最大执行秒数。
	TaskTimeoutSeconds int `json:"taskTimeoutSeconds" gorm:"not null;default:300"`
	// LastStatus 是最近一次健康检查状态。
	LastStatus string `json:"lastStatus" gorm:"size:32;not null;default:unknown"`
	// LastCheckedAt 是最近一次健康检查时间。
	LastCheckedAt *time.Time `json:"lastCheckedAt"`
	// LastError 是最近一次健康检查错误。
	LastError string `json:"lastError" gorm:"size:1000"`
	// HasAuthToken 表示配置是否已有认证令牌，仅用于响应投影。
	HasAuthToken bool `json:"hasAuthToken" gorm:"-"`
}

// TableName 返回 MCP 服务配置表名。
func (MCPServer) TableName() string { return "xy_mcp_servers" }

// AgentConfig 定义 Agent 单次运行的资源与安全上限。
type AgentConfig struct {
	// ID 是全局配置的固定主键。
	ID uint `json:"ID" gorm:"primaryKey"`
	// MaxWallTimeSeconds 是单次运行最大墙钟秒数。
	MaxWallTimeSeconds int `json:"maxWallTimeSeconds" gorm:"not null"`
	// MaxIterations 是推理循环最大轮数。
	MaxIterations int `json:"maxIterations" gorm:"not null"`
	// MaxModelCalls 是单次运行最大模型调用数。
	MaxModelCalls int `json:"maxModelCalls" gorm:"not null"`
	// MaxToolCalls 是单次运行最大工具调用数。
	MaxToolCalls int `json:"maxToolCalls" gorm:"not null"`
	// MaxConcurrentToolCalls 是并行工具调用上限。
	MaxConcurrentToolCalls int `json:"maxConcurrentToolCalls" gorm:"not null;default:3"`
	// MaxSearchCalls 是搜索类工具调用上限。
	MaxSearchCalls int `json:"maxSearchCalls" gorm:"not null"`
	// MaxWriteToolCalls 是写操作工具调用上限。
	MaxWriteToolCalls int `json:"maxWriteToolCalls" gorm:"not null"`
	// MaxOutputTokens 是单次模型输出令牌上限。
	MaxOutputTokens int `json:"maxOutputTokens" gorm:"not null"`
	// ConversationHistoryLimit 是加载历史消息的最大条数。
	ConversationHistoryLimit int `json:"conversationHistoryLimit" gorm:"not null"`
	// SummaryTriggerTokens 是触发会话摘要的令牌阈值。
	SummaryTriggerTokens int `json:"summaryTriggerTokens" gorm:"not null"`
}

// TableName 返回 Agent 全局配置表名。
func (AgentConfig) TableName() string { return "xy_agent_config" }

// DefaultAgentConfig 返回服务首次启动使用的保守运行上限。
func DefaultAgentConfig() AgentConfig {
	return AgentConfig{
		ID: 1, MaxWallTimeSeconds: 300, MaxIterations: 50, MaxModelCalls: 60,
		MaxToolCalls: 200, MaxConcurrentToolCalls: 3, MaxSearchCalls: 200, MaxWriteToolCalls: 10,
		MaxOutputTokens: 2000, ConversationHistoryLimit: 40, SummaryTriggerTokens: 24000,
	}
}
