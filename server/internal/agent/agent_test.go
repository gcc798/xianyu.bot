package agent

import (
	"context"
	"testing"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestInitializeSeedsAgentConfig(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentConfig{}))

	require.NoError(t, Initialize(t.Context(), db))
	var config model.AgentConfig
	require.NoError(t, db.First(&config, 1).Error)
}

func TestInitializeRejectsMissingConnection(t *testing.T) {
	require.ErrorContains(t, Initialize(t.Context(), nil), "database is not initialized")
}

func TestNewRunnerUsesRootContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	runner := NewRunner(ctx, nil, nil, nil, nil, nil)
	require.Equal(t, ctx, runner.root)
}
