package request

type SaveAIModel struct {
	Name                  string `json:"name" binding:"required"`
	Provider              string `json:"provider" binding:"required"`
	BaseURL               string `json:"baseUrl" binding:"required"`
	APIKey                string `json:"apiKey"`
	Model                 string `json:"model" binding:"required"`
	Enabled               bool   `json:"enabled"`
	IsDefault             bool   `json:"isDefault"`
	SupportsTools         bool   `json:"supportsTools"`
	SupportsStreaming     bool   `json:"supportsStreaming"`
	ContextWindow         int    `json:"contextWindow"`
	MaxOutputTokens       int    `json:"maxOutputTokens"`
	RequestTimeoutSeconds int    `json:"requestTimeoutSeconds"`
	ExtraConfig           string `json:"extraConfig"`
}

type SaveMCPServer struct {
	Name                  string `json:"name" binding:"required"`
	Code                  string `json:"code" binding:"required"`
	BaseURL               string `json:"baseUrl" binding:"required"`
	AuthHeader            string `json:"authHeader"`
	AuthToken             string `json:"authToken"`
	Enabled               bool   `json:"enabled"`
	RequestTimeoutSeconds int    `json:"requestTimeoutSeconds"`
	TaskTimeoutSeconds    int    `json:"taskTimeoutSeconds"`
}
