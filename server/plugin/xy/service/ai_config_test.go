package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model"
	xyRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAIConfigUpdatePreservesSecrets(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AIModel{}, &model.MCPServer{}, &model.AgentConfig{}))
	oldDB := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = oldDB })

	service := new(aiConfig)
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

func TestAIConfigRejectsUnusableDefaultModel(t *testing.T) {
	input := validModelInput()
	input.IsDefault = true
	input.SupportsTools = false
	_, err := buildAIModel(input)
	require.ErrorContains(t, err, "支持工具调用")

	input.SupportsTools = true
	input.Enabled = false
	_, err = buildAIModel(input)
	require.ErrorContains(t, err, "必须已启用")
}

func TestMCPCodeValidation(t *testing.T) {
	input := validMCPInput()
	input.Code = "Telegram Resource"
	_, err := buildMCPServer(input)
	require.ErrorContains(t, err, "编码必须")

	input.Code = strings.Repeat("a", 65)
	_, err = buildMCPServer(input)
	require.ErrorContains(t, err, "编码必须")
}

func validModelInput() xyRequest.SaveAIModel {
	return xyRequest.SaveAIModel{
		Name: "OpenAI", Provider: "openai", BaseURL: "https://api.openai.com/v1", Model: "gpt-4.1",
		Enabled: true, SupportsTools: true, SupportsStreaming: true, ContextWindow: 128000,
		MaxOutputTokens: 2000, RequestTimeoutSeconds: 60,
	}
}

func validMCPInput() xyRequest.SaveMCPServer {
	return xyRequest.SaveMCPServer{
		Name: "Telegram", Code: "telegram", BaseURL: "http://telegram:8080/mcp", Enabled: true,
		RequestTimeoutSeconds: 15, TaskTimeoutSeconds: 300,
	}
}
