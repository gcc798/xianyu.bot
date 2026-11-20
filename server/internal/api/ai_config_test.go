package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gcc798/xianyu.bot/internal/agent"
	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestAIModelConfigIncludesPlainAPIKey 验证管理员模型配置投影包含完整凭据。
func TestAIModelConfigIncludesPlainAPIKey(t *testing.T) {
	payload, err := json.Marshal(aiModelConfig{
		AIModel: model.AIModel{Name: "OpenAI", APIKey: "plain-secret", HasAPIKey: true},
		APIKey:  "plain-secret",
	})
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(payload, &decoded))
	require.Equal(t, "plain-secret", decoded["apiKey"])
}

// TestMCPServerProjectionExcludesCredential 验证 MCP 管理投影包含策略但不返回凭据。
func TestMCPServerProjectionExcludesCredential(t *testing.T) {
	payload, err := json.Marshal(model.MCPServer{
		Name: "资源服务", AuthToken: "plain-secret", HasAuthToken: true,
		Tags: []string{"资源"}, EnabledTools: []string{"search"}, ConnectionScope: "owner",
	})
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(payload, &decoded))
	require.NotContains(t, string(payload), "plain-secret")
	require.Equal(t, true, decoded["hasAuthToken"])
	require.Equal(t, "owner", decoded["connectionScope"])
	require.Equal(t, []any{"资源"}, decoded["tags"])
}

// TestListMCPServerToolsUsesEnvelopeAndHidesSchema 验证缓存工具接口信封与敏感字段投影。
func TestListMCPServerToolsUsesEnvelopeAndHidesSchema(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.MCPServer{}, &model.MCPTool{}))
	server := model.MCPServer{Name: "资源", Code: "resources", BaseURL: "http://resources/mcp", AuthToken: "server-secret", Enabled: true}
	require.NoError(t, db.Create(&server).Error)
	tool := model.MCPTool{
		MCPServerCode: server.Code, Name: "search", Description: "搜索", InputSchema: `{"type":"object","default":"schema-secret"}`,
		Definition: `{"name":"search","inputSchema":{"type":"object"}}`, SearchText: "搜索 search",
		DefinitionHash: "hash", DiscoveredAt: time.Now(),
	}
	require.NoError(t, db.Create(&tool).Error)
	registry := agent.NewMCPRegistry(db)
	require.NoError(t, registry.LoadCatalog(t.Context()))
	handler := NewAIConfigAPI(agent.NewAIConfig(db, registry))

	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(server.ID), 10)}}
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ListMCPServerTools(ctx)

	require.Equal(t, http.StatusOK, response.Code)
	var envelope struct {
		Code int `json:"code"`
		Data struct {
			List []map[string]any `json:"list"`
		} `json:"data"`
		Msg string `json:"msg"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &envelope))
	require.Zero(t, envelope.Code)
	require.Equal(t, "成功", envelope.Msg)
	require.Len(t, envelope.Data.List, 1)
	require.Equal(t, "search", envelope.Data.List[0]["name"])
	require.NotContains(t, response.Body.String(), "server-secret")
	require.NotContains(t, response.Body.String(), "schema-secret")
	require.NotContains(t, response.Body.String(), "inputSchema")
}

// TestMCPActionEndpointsUseFailureEnvelope 验证健康检查与目录刷新沿用统一错误信封。
func TestMCPActionEndpointsUseFailureEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAIConfigAPI(nil)
	for _, test := range []struct {
		name   string
		handle func(*gin.Context)
	}{
		{name: "health", handle: handler.CheckMCPServerHealth},
		{name: "catalog", handle: handler.RefreshMCPServerCatalog},
		{name: "tools", handle: handler.ListMCPServerTools},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(response)
			ctx.Params = gin.Params{{Key: "id", Value: "invalid"}}
			ctx.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			test.handle(ctx)
			var envelope struct {
				Code int    `json:"code"`
				Msg  string `json:"msg"`
			}
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &envelope))
			require.Equal(t, 7, envelope.Code)
			require.Equal(t, "ID 无效", envelope.Msg)
		})
	}
}
