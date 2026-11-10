package agent

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestAIConfigUpdatePreservesSecrets 验证对应测试场景。
func TestAIConfigUpdatePreservesSecrets(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AIModel{}, &model.MCPServer{}, &model.AgentConfig{}))

	service := &AIConfig{db: db}
	modelInput := validModelInput()
	modelInput.APIKey = "secret-model-key"
	createdModel, err := service.CreateModel(t.Context(), modelInput)
	require.NoError(t, err)
	modelInput.APIKey = ""
	modelInput.Name = "updated"
	require.NoError(t, service.UpdateModel(t.Context(), createdModel.ID, modelInput))

	serverInput := validMCPInput()
	serverInput.AuthToken = "secret-mcp-token"
	createdServer, err := service.CreateMCPServer(t.Context(), serverInput)
	require.NoError(t, err)
	serverInput.AuthToken = ""
	serverInput.Name = "updated"
	require.NoError(t, service.UpdateMCPServer(t.Context(), createdServer.ID, serverInput))

	var storedModel model.AIModel
	require.NoError(t, db.First(&storedModel, createdModel.ID).Error)
	require.Equal(t, "secret-model-key", storedModel.APIKey)
	var storedServer model.MCPServer
	require.NoError(t, db.First(&storedServer, createdServer.ID).Error)
	require.Equal(t, "secret-mcp-token", storedServer.AuthToken)

	models, err := service.ListModels(t.Context())
	require.NoError(t, err)
	servers, err := service.ListMCPServers(t.Context())
	require.NoError(t, err)
	payload, err := json.Marshal(struct {
		Models  []model.AIModel   `json:"models"`
		Servers []model.MCPServer `json:"servers"`
	}{Models: models, Servers: servers})
	require.NoError(t, err)
	require.NotContains(t, string(payload), "secret-")
	require.True(t, models[0].HasAPIKey)
	require.True(t, servers[0].HasAuthToken)
}

// TestAIConfigRejectsInvalidPriority 验证模型优先级必须为正整数。
func TestAIConfigRejectsInvalidPriority(t *testing.T) {
	input := validModelInput()
	input.Priority = 0
	_, err := buildAIModel(input)
	require.ErrorContains(t, err, "优先级")
}

// TestMCPCodeValidation 验证对应测试场景。
func TestMCPCodeValidation(t *testing.T) {
	input := validMCPInput()
	input.Code = "Telegram Resource"
	_, err := buildMCPServer(input)
	require.ErrorContains(t, err, "编码必须")

	input.Code = strings.Repeat("a", 65)
	_, err = buildMCPServer(input)
	require.ErrorContains(t, err, "编码必须")
}

func TestMCPServerAcceptsDockerServiceURL(t *testing.T) {
	input := validMCPInput()
	input.BaseURL = "http://telegram:8080/mcp/"
	_, err := buildMCPServer(input)
	require.NoError(t, err)
}

// validModelInput 提供测试辅助能力。
func validModelInput() xyRequest.SaveAIModel {
	return xyRequest.SaveAIModel{
		Name: "OpenAI", Provider: "openai", BaseURL: "https://api.openai.com/v1", Model: "gpt-4.1",
		Enabled: true, Priority: 1, SupportsTools: true, SupportsStreaming: true, ContextWindow: 128000,
		MaxOutputTokens: 2000, RequestTimeoutSeconds: 60,
	}
}

// validMCPInput 提供测试辅助能力。
func validMCPInput() xyRequest.SaveMCPServer {
	return xyRequest.SaveMCPServer{
		Name: "Telegram", Code: "telegram", BaseURL: "http://telegram:8080/mcp", Enabled: true,
		RequestTimeoutSeconds: 15, TaskTimeoutSeconds: 300,
	}
}
