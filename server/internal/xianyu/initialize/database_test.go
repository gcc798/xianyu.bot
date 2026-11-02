package initialize

import (
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestDatabaseSeedsAgentConfig(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentConfig{}))

	oldDB := global.GVA_DB
	t.Cleanup(func() { global.GVA_DB = oldDB })
	global.GVA_DB = db

	require.NoError(t, Database(t.Context()))
	var config model.AgentConfig
	require.NoError(t, db.First(&config, 1).Error)
}

func TestDatabaseRejectsMissingConnection(t *testing.T) {
	oldDB := global.GVA_DB
	t.Cleanup(func() { global.GVA_DB = oldDB })
	global.GVA_DB = nil
	require.ErrorContains(t, Database(t.Context()), "database is not initialized")
}
