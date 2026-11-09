package api

import (
	"encoding/json"
	"testing"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/stretchr/testify/require"
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
