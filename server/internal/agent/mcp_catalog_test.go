package agent

import (
	"testing"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestMCPCatalogLoadsSnapshotAndAppliesPolicy 验证重启恢复、策略过滤和坏工具隔离。
func TestMCPCatalogLoadsSnapshotAndAppliesPolicy(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.MCPServer{}, &model.MCPTool{}))
	require.NoError(t, db.Create(&[]model.MCPServer{
		{Name: "资源", Code: "resources", BaseURL: "http://resources/mcp", Enabled: true, EnabledTools: []string{"keep", "blocked", "invalid"}, DisabledTools: []string{"blocked"}},
		{Name: "停用", Code: "disabled", BaseURL: "http://disabled/mcp", Enabled: false},
	}).Error)
	require.NoError(t, db.Model(&model.MCPServer{}).Where("code = ?", "disabled").Update("enabled", false).Error)
	now := time.Now().UTC()
	require.NoError(t, db.Create(&[]model.MCPTool{
		catalogModelTool("resources", "keep", `{"name":"keep","inputSchema":{"type":"object","properties":{}}}`, now),
		catalogModelTool("resources", "blocked", `{"name":"blocked","inputSchema":{"type":"object"}}`, now),
		catalogModelTool("resources", "other", `{"name":"other","inputSchema":{"type":"object"}}`, now),
		catalogModelTool("resources", "invalid", `{`, now),
		catalogModelTool("disabled", "hidden", `{"name":"hidden","inputSchema":{"type":"object"}}`, now),
	}).Error)

	registry := NewMCPRegistry(db)
	require.NoError(t, registry.LoadCatalog(t.Context()))
	tools := registry.CatalogTools()
	require.Len(t, tools, 1)
	require.Equal(t, "resources__keep", tools[0].Code())
	require.Len(t, registry.CachedTools("resources"), 1)

	tools[0].Snapshot.Name = "mutated"
	tools[0].Remote.InputSchema.Properties["mutated"] = true
	loaded, ok := registry.CatalogTool("resources__keep")
	require.True(t, ok)
	require.Equal(t, "keep", loaded.Snapshot.Name)
	require.NotContains(t, loaded.Remote.InputSchema.Properties, "mutated")

	restarted := NewMCPRegistry(db)
	require.NoError(t, restarted.LoadCatalog(t.Context()))
	_, ok = restarted.CatalogTool("resources__keep")
	require.True(t, ok)
}

// TestBuildMCPCatalogKeepsFirstDuplicateCode 验证重复工具编码不会覆盖已发布定义。
func TestBuildMCPCatalogKeepsFirstDuplicateCode(t *testing.T) {
	server := model.MCPServer{Name: "资源", Code: "resources", Enabled: true}
	first := catalogModelTool("resources", "search", `{"name":"search","description":"first","inputSchema":{"type":"object"}}`, time.Now())
	second := first
	second.ID = 2
	second.Definition = `{"name":"search","description":"second","inputSchema":{"type":"object"}}`

	snapshot := buildMCPCatalog([]model.MCPServer{server}, []model.MCPTool{first, second})
	require.Len(t, snapshot.tools, 1)
	require.Equal(t, "first", snapshot.tools[0].Remote.Description)
}

// catalogModelTool 创建目录测试记录。
func catalogModelTool(serverCode, name, definition string, discoveredAt time.Time) model.MCPTool {
	return model.MCPTool{
		MCPServerCode: serverCode, Name: name, InputSchema: `{}`, Definition: definition,
		SearchText: name, DefinitionHash: name + "-hash", DiscoveredAt: discoveredAt,
	}
}
