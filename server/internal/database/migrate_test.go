package database

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestMigratePostgres(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, Migrate(t.Context(), db, "pgsql"))

	var removed, retained *string
	require.NoError(t, db.Raw("SELECT to_regclass('public.sys_apis')::text").Scan(&removed).Error)
	require.Nil(t, removed)
	require.NoError(t, db.Raw("SELECT to_regclass('public.sys_users')::text").Scan(&retained).Error)
	require.NotNil(t, retained)
	var userSettingColumns int64
	require.NoError(t, db.Raw(`SELECT count(*) FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'sys_users' AND column_name = 'origin_setting'`).Scan(&userSettingColumns).Error)
	require.Zero(t, userSettingColumns)
}

func TestMigrateRejectsUnsupportedDatabase(t *testing.T) {
	err := Migrate(context.Background(), &gorm.DB{}, "sqlite")
	require.ErrorContains(t, err, `unsupported database type "sqlite"`)
}
