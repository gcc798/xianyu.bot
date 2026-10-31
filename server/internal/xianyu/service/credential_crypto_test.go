package service

import (
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCredentialKeyFailsWithoutEnvironmentOrDatabase(t *testing.T) {
	t.Setenv(credentialKeyEnvironment, "")
	oldDB := global.GVA_DB
	t.Cleanup(func() { global.GVA_DB = oldDB })
	global.GVA_DB = nil

	_, err := credentialKey()
	require.Error(t, err)
}

func TestValidateCredentialKeyInitializesEmptyDatabase(t *testing.T) {
	t.Setenv(credentialKeyEnvironment, "")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Setting{}))

	oldDB := global.GVA_DB
	t.Cleanup(func() { global.GVA_DB = oldDB })
	global.GVA_DB = db

	require.NoError(t, ValidateCredentialKey())
	first, err := credentialKey()
	require.NoError(t, err)
	require.Len(t, first, 32)

	require.NoError(t, ValidateCredentialKey())
	second, err := credentialKey()
	require.NoError(t, err)
	require.Equal(t, first, second)
}
