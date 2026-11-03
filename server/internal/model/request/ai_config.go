package request

// SaveAIModel 描述创建或更新模型配置的输入。
type SaveAIModel struct {
	// Name 是配置显示名称。
	Name string `json:"name" binding:"required"`
	// Provider 标识模型服务商或兼容协议。
	Provider string `json:"provider" binding:"required"`
	// BaseURL 是模型 API 基础地址。
	BaseURL string `json:"baseUrl" binding:"required"`
	// APIKey 是新的模型服务凭据，空值表示不替换。
	APIKey string `json:"apiKey"`
	// Model 是服务商侧模型标识。
	Model string `json:"model" binding:"required"`
	// Enabled 决定配置是否可被选择。
	Enabled bool `json:"enabled"`
	// IsDefault 决定是否作为默认模型。
	IsDefault bool `json:"isDefault"`
	// SupportsTools 表示模型是否支持工具调用。
	SupportsTools bool `json:"supportsTools"`
	// SupportsStreaming 表示模型是否支持流式输出。
	SupportsStreaming bool `json:"supportsStreaming"`
	// ContextWindow 是模型最大上下文令牌数。
	ContextWindow int `json:"contextWindow"`
	// MaxOutputTokens 是单次生成的最大令牌数。
	MaxOutputTokens int `json:"maxOutputTokens"`
	// RequestTimeoutSeconds 是单次模型请求超时秒数。
	RequestTimeoutSeconds int `json:"requestTimeoutSeconds"`
	// ExtraConfig 是服务商专属的 JSON 配置。
	ExtraConfig string `json:"extraConfig"`
}

// SaveMCPServer 描述创建或更新 MCP 服务配置的输入。
type SaveMCPServer struct {
	// Name 是 MCP 服务显示名称。
	Name string `json:"name" binding:"required"`
	// Code 是 MCP 服务稳定唯一编码。
	Code string `json:"code" binding:"required"`
	// BaseURL 是 MCP HTTP 入口地址。
	BaseURL string `json:"baseUrl" binding:"required"`
	// AuthHeader 是认证令牌使用的请求头名称。
	AuthHeader string `json:"authHeader"`
	// AuthToken 是新的认证凭据，空值表示不替换。
	AuthToken string `json:"authToken"`
	// Enabled 决定服务及其工具是否可用。
	Enabled bool `json:"enabled"`
	// RequestTimeoutSeconds 是普通请求超时秒数。
	RequestTimeoutSeconds int `json:"requestTimeoutSeconds"`
	// TaskTimeoutSeconds 是长任务最大执行秒数。
	TaskTimeoutSeconds int `json:"taskTimeoutSeconds"`
}
