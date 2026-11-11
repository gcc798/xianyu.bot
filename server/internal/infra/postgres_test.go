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
	require.Equal(t, int64(202611100008), version)
	var hasJWTBlacklist bool
	require.NoError(t, db.Raw("SELECT to_regclass('public.jwt_blacklists') IS NOT NULL").Scan(&hasJWTBlacklist).Error)
	require.False(t, hasJWTBlacklist)
	var hasUpdatedAt bool
	require.NoError(t, db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'tel_resource_tasks' AND column_name = 'updated_at'
	)`).Scan(&hasUpdatedAt).Error)
	require.True(t, hasUpdatedAt)
	var hasFailureDetails bool
	require.NoError(t, db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'tel_resource_tasks' AND column_name = 'failure_details'
	)`).Scan(&hasFailureDetails).Error)
	require.True(t, hasFailureDetails)
	var orderStatusCount int64
	require.NoError(t, db.Raw(`SELECT count(*) FROM s_dict_detail d
		JOIN s_dict s ON s.id = d.sys_dictionary_id
		WHERE s.type = 'order_status' AND d.status IS TRUE`).Scan(&orderStatusCount).Error)
	require.Equal(t, int64(3), orderStatusCount)
}
