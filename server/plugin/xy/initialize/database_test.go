package initialize

import (
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestDatabaseConfiguresSQLite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

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
	for _, table := range []string{"xy_resource_searches", "xy_resource_candidates", "xy_resource_tasks", "xy_resources", "xy_resource_sources"} {
		require.Truef(t, db.Migrator().HasTable(table), "expected resource table %s", table)
	}
	for _, table := range []string{"xy_agent_sessions", "xy_agent_messages", "xy_agent_runs", "xy_agent_run_tools"} {
		require.Truef(t, db.Migrator().HasTable(table), "expected agent table %s", table)
	}
}

func TestDatabaseSkipsSQLitePragmasForPostgres(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

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
