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
}

func TestMigrateRejectsUnsupportedDatabase(t *testing.T) {
	err := Migrate(context.Background(), &gorm.DB{}, "sqlite")
	require.ErrorContains(t, err, `unsupported database type "sqlite"`)
}
