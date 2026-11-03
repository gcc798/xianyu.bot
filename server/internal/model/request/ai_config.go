package request

// SaveAIModel 描述创建或更新模型配置的输入。
type SaveAIModel struct {
	Name                  string `json:"name" binding:"required"`     // Name 是配置显示名称。
	Provider              string `json:"provider" binding:"required"` // Provider 标识模型服务商或兼容协议。
	BaseURL               string `json:"baseUrl" binding:"required"`  // BaseURL 是模型 API 基础地址。
	APIKey                string `json:"apiKey"`                      // APIKey 是新的模型服务凭据，空值表示不替换。
	Model                 string `json:"model" binding:"required"`    // Model 是服务商侧模型标识。
	Enabled               bool   `json:"enabled"`                     // Enabled 决定配置是否可被选择。
	IsDefault             bool   `json:"isDefault"`                   // IsDefault 决定是否作为默认模型。
	SupportsTools         bool   `json:"supportsTools"`               // SupportsTools 表示模型是否支持工具调用。
	SupportsStreaming     bool   `json:"supportsStreaming"`           // SupportsStreaming 表示模型是否支持流式输出。
	ContextWindow         int    `json:"contextWindow"`               // ContextWindow 是模型最大上下文令牌数。
	MaxOutputTokens       int    `json:"maxOutputTokens"`             // MaxOutputTokens 是单次生成的最大令牌数。
	RequestTimeoutSeconds int    `json:"requestTimeoutSeconds"`       // RequestTimeoutSeconds 是单次模型请求超时秒数。
	ExtraConfig           string `json:"extraConfig"`                 // ExtraConfig 是服务商专属的 JSON 配置。
}

// SaveMCPServer 描述创建或更新 MCP 服务配置的输入。
type SaveMCPServer struct {
	Name                  string `json:"name" binding:"required"`    // Name 是 MCP 服务显示名称。
	Code                  string `json:"code" binding:"required"`    // Code 是 MCP 服务稳定唯一编码。
	BaseURL               string `json:"baseUrl" binding:"required"` // BaseURL 是 MCP HTTP 入口地址。
	AuthHeader            string `json:"authHeader"`                 // AuthHeader 是认证令牌使用的请求头名称。
	AuthToken             string `json:"authToken"`                  // AuthToken 是新的认证凭据，空值表示不替换。
	Enabled               bool   `json:"enabled"`                    // Enabled 决定服务及其工具是否可用。
	RequestTimeoutSeconds int    `json:"requestTimeoutSeconds"`      // RequestTimeoutSeconds 是普通请求超时秒数。
	TaskTimeoutSeconds    int    `json:"taskTimeoutSeconds"`         // TaskTimeoutSeconds 是长任务最大执行秒数。
}
