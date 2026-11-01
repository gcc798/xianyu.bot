package initialize

import (
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestDatabaseConfiguresSQLite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentConfig{}))

	oldDB := global.GVA_DB
	oldType := global.GVA_CONFIG.System.DbType
	t.Cleanup(func() {
		global.GVA_DB = oldDB
		global.GVA_CONFIG.System.DbType = oldType
	})
	global.GVA_DB = db
	global.GVA_CONFIG.System.DbType = "sqlite"

	require.NoError(t, Database(t.Context()))

	var foreignKeys int
	require.NoError(t, db.Raw("PRAGMA foreign_keys").Scan(&foreignKeys).Error)
	require.Equal(t, 1, foreignKeys)
	var config model.AgentConfig
	require.NoError(t, db.First(&config, 1).Error)
}

func TestDatabaseSkipsSQLitePragmasForPostgres(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentConfig{}))

	oldDB := global.GVA_DB
	oldType := global.GVA_CONFIG.System.DbType
	t.Cleanup(func() {
		global.GVA_DB = oldDB
		global.GVA_CONFIG.System.DbType = oldType
	})
	global.GVA_DB = db
	global.GVA_CONFIG.System.DbType = "pgsql"

	require.NoError(t, Database(t.Context()))
}
