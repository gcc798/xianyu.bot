package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestMCPToolPersistsJSONAndUniqueCode 验证工具定义 JSON 往返和服务内名称唯一。
func TestMCPToolPersistsJSONAndUniqueCode(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&MCPTool{}))

	tool := MCPTool{
		MCPServerCode: "resources", Name: "search", Description: "搜索资源",
		InputSchema:  `{"type":"object","properties":{"token":{"default":"secret-token"}}}`,
		OutputSchema: `{"type":"object"}`, Annotations: `{"readOnlyHint":true}`,
		Definition: `{"name":"search"}`, SearchText: "资源 搜索 search",
		DefinitionHash: "0123456789abcdef", DiscoveredAt: time.Now().UTC(),
	}
	require.NoError(t, db.Create(&tool).Error)

	var stored MCPTool
	require.NoError(t, db.First(&stored, tool.ID).Error)
	require.JSONEq(t, tool.InputSchema, stored.InputSchema)
	require.JSONEq(t, tool.OutputSchema, stored.OutputSchema)
	require.JSONEq(t, tool.Annotations, stored.Annotations)
	require.JSONEq(t, tool.Definition, stored.Definition)

	duplicate := tool
	duplicate.ID = 0
	require.Error(t, db.Create(&duplicate).Error)

	payload, err := json.Marshal(stored)
	require.NoError(t, err)
	require.NotContains(t, string(payload), "secret-token")
	require.NotContains(t, string(payload), `"inputSchema"`)
	require.NotContains(t, string(payload), `"definition"`)
}
