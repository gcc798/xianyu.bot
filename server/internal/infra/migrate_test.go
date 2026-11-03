package infra

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestMigratePostgres 验证迁移能清理废弃表并保留当前业务结构。
func TestMigratePostgres(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, Migrate(t.Context(), db))

	var removed, retained *string
	require.NoError(t, db.Raw("SELECT to_regclass('public.sys_apis')::text").Scan(&removed).Error)
	require.Nil(t, removed)
	require.NoError(t, db.Raw("SELECT to_regclass('public.sys_users')::text").Scan(&retained).Error)
	require.NotNil(t, retained)
	var userSettingColumns int64
	require.NoError(t, db.Raw(`SELECT count(*) FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'sys_users' AND column_name = 'origin_setting'`).Scan(&userSettingColumns).Error)
	require.Zero(t, userSettingColumns)
	var toolIdentityColumns int64
	require.NoError(t, db.Raw(`SELECT count(*) FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'xy_agent_run_tools' AND column_name IN ('tool_code', 'tool_source')`).Scan(&toolIdentityColumns).Error)
	require.EqualValues(t, 2, toolIdentityColumns)
}

// TestMigrateRejectsUninitializedDatabase 验证未初始化数据库会被明确拒绝。
func TestMigrateRejectsUninitializedDatabase(t *testing.T) {
	err := Migrate(t.Context(), nil)
	require.ErrorContains(t, err, "database is not initialized")
}
