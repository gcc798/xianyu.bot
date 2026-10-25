package service

import (
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/stretchr/testify/require"
)

func TestCredentialKeyFailsWithoutEnvironmentOrDatabase(t *testing.T) {
	t.Setenv(credentialKeyEnvironment, "")
	oldDB := global.GVA_DB
	t.Cleanup(func() { global.GVA_DB = oldDB })
	global.GVA_DB = nil

	_, err := credentialKey()
	require.Error(t, err)
}
