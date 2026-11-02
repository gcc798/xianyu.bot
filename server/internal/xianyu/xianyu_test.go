package xianyu

import (
	"testing"

	"github.com/gcc798/go.xianyu.assistant/server/internal/infra"
	"github.com/gcc798/go.xianyu.assistant/server/internal/model/xianyu"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestDatabaseSeedsAgentConfig(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentConfig{}))

	oldDB := infra.DB
	t.Cleanup(func() { infra.DB = oldDB })
	infra.DB = db

	require.NoError(t, initializeDatabase(t.Context()))
	var config model.AgentConfig
	require.NoError(t, db.First(&config, 1).Error)
}

func TestDatabaseRejectsMissingConnection(t *testing.T) {
	oldDB := infra.DB
	t.Cleanup(func() { infra.DB = oldDB })
	infra.DB = nil
	require.ErrorContains(t, initializeDatabase(t.Context()), "database is not initialized")
}
