package service

import (
	"testing"

	"github.com/gcc798/go.xianyu.assistant/server/internal/infra"
	"github.com/gcc798/go.xianyu.assistant/server/internal/model/xianyu"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCredentialKeyFailsWithoutEnvironmentOrDatabase(t *testing.T) {
	t.Setenv(credentialKeyEnvironment, "")
	oldDB := infra.DB
	t.Cleanup(func() { infra.DB = oldDB })
	infra.DB = nil

	_, err := credentialKey()
	require.Error(t, err)
}

func TestValidateCredentialKeyInitializesEmptyDatabase(t *testing.T) {
	t.Setenv(credentialKeyEnvironment, "")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Setting{}))

	oldDB := infra.DB
	t.Cleanup(func() { infra.DB = oldDB })
	infra.DB = db

	require.NoError(t, ValidateCredentialKey())
	first, err := credentialKey()
	require.NoError(t, err)
	require.Len(t, first, 32)

	require.NoError(t, ValidateCredentialKey())
	second, err := credentialKey()
	require.NoError(t, err)
	require.Equal(t, first, second)
}
