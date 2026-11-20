package infra

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestMigratePostgresInitializesEmptyDatabase(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN is required")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })

	var hasObjects bool
	require.NoError(t, db.Raw(`SELECT EXISTS (
		SELECT 1 FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'public' AND c.relkind IN ('r', 'p', 'v', 'm', 'S', 'f')
	)`).Scan(&hasObjects).Error)
	require.False(t, hasObjects, "TEST_POSTGRES_DSN must point to an empty database")

	require.NoError(t, MigratePostgres(t.Context(), db))
	require.NoError(t, MigratePostgres(t.Context(), db))

	var version int64
	require.NoError(t, db.Raw(`SELECT max(version_id) FROM goose_db_version WHERE is_applied`).Scan(&version).Error)
	require.Equal(t, int64(202611200010), version)
	var renamedTableCount int
	require.NoError(t, db.Raw(`SELECT count(*) FROM unnest(ARRAY[
		'agent_config', 'agent_events', 'agent_messages', 'agent_run_tools', 'agent_runs', 'agent_sessions', 'agent_models',
		'mcp_servers', 'mcp_tools',
		'resource_candidates', 'resource_searches', 'resource_sources', 'resource_tasks', 'resources',
		'tg_api_configs', 'tg_acquisition_tasks', 'tg_dialogs', 'tg_leave_history', 'tg_resource_files',
		'tg_resource_sources', 'tg_resource_tasks', 'tg_scan_history', 'tg_settings'
	]) AS table_name WHERE to_regclass('public.' || table_name) IS NOT NULL`).Scan(&renamedTableCount).Error)
	require.Equal(t, 23, renamedTableCount)
	var hasCatalogHash bool
	require.NoError(t, db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'agent_run_tools' AND column_name = 'catalog_hash'
	)`).Scan(&hasCatalogHash).Error)
	require.True(t, hasCatalogHash)
	var hasJWTBlacklist bool
	require.NoError(t, db.Raw("SELECT to_regclass('public.jwt_blacklists') IS NOT NULL").Scan(&hasJWTBlacklist).Error)
	require.False(t, hasJWTBlacklist)
	var hasUpdatedAt bool
	require.NoError(t, db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'tg_resource_tasks' AND column_name = 'updated_at'
	)`).Scan(&hasUpdatedAt).Error)
	require.True(t, hasUpdatedAt)
	var hasFailureDetails bool
	require.NoError(t, db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'tg_resource_tasks' AND column_name = 'failure_details'
	)`).Scan(&hasFailureDetails).Error)
	require.True(t, hasFailureDetails)
	var orderStatusCount int64
	require.NoError(t, db.Raw(`SELECT count(*) FROM s_dict_detail d
		JOIN s_dict s ON s.id = d.sys_dictionary_id
		WHERE s.type = 'order_status' AND d.status IS TRUE`).Scan(&orderStatusCount).Error)
	require.Equal(t, int64(3), orderStatusCount)
}

func TestMigratePostgresRenamesLegacyDomainTables(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_UPGRADE_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_UPGRADE_DSN is required")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })

	var version int64
	require.NoError(t, db.Raw(`SELECT max(version_id) FROM goose_db_version WHERE is_applied`).Scan(&version).Error)
	require.Equal(t, int64(202611200009), version, "upgrade DSN must contain the legacy schema")
	require.NoError(t, MigratePostgres(t.Context(), db))
	require.NoError(t, db.Raw(`SELECT max(version_id) FROM goose_db_version WHERE is_applied`).Scan(&version).Error)
	require.Equal(t, int64(202611200010), version)

	var renamedTableCount int
	require.NoError(t, db.Raw(`SELECT count(*) FROM unnest(ARRAY[
		'agent_config', 'agent_events', 'agent_messages', 'agent_run_tools', 'agent_runs', 'agent_sessions', 'agent_models',
		'mcp_servers', 'mcp_tools',
		'resource_candidates', 'resource_searches', 'resource_sources', 'resource_tasks', 'resources',
		'tg_api_configs', 'tg_acquisition_tasks', 'tg_dialogs', 'tg_leave_history', 'tg_resource_files',
		'tg_resource_sources', 'tg_resource_tasks', 'tg_scan_history', 'tg_settings'
	]) AS table_name WHERE to_regclass('public.' || table_name) IS NOT NULL`).Scan(&renamedTableCount).Error)
	require.Equal(t, 23, renamedTableCount)
}
