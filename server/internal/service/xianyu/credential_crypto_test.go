package service

import (
	"testing"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCredentialKeyFailsWithoutEnvironmentOrDatabase(t *testing.T) {
	t.Setenv(credentialKeyEnvironment, "")
	_, err := credentialKey(nil)
	require.Error(t, err)
}

func TestValidateCredentialKeyInitializesEmptyDatabase(t *testing.T) {
	t.Setenv(credentialKeyEnvironment, "")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Setting{}))

	require.NoError(t, ValidateCredentialKey(db))
	first, err := credentialKey(db)
	require.NoError(t, err)
	require.Len(t, first, 32)

	require.NoError(t, ValidateCredentialKey(db))
	second, err := credentialKey(db)
	require.NoError(t, err)
	require.Equal(t, first, second)
}
