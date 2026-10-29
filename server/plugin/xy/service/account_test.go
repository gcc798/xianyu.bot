package service

import (
	"encoding/base64"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model"
	xyRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAccountOwnerIsolationAndCredentialEncryption(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Credential{}))

	oldDB := global.GVA_DB
	t.Cleanup(func() { global.GVA_DB = oldDB })
	global.GVA_DB = db
	t.Setenv(credentialKeyEnvironment, base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))

	service := new(account)
	created, err := service.Create(t.Context(), 10, xyRequest.CreateAccount{Label: "主账号"})
	require.NoError(t, err)

	require.Error(t, service.Update(t.Context(), 11, xyRequest.UpdateAccount{ID: created.ID, Label: "越权修改"}))
	require.Error(t, service.SaveCredential(t.Context(), 11, xyRequest.SaveCredential{ID: created.ID, Cookie: "secret-cookie"}))
	require.NoError(t, service.SaveCredential(t.Context(), 10, xyRequest.SaveCredential{ID: created.ID, Cookie: "secret-cookie", Token: "secret-token"}))

	var credential model.Credential
	require.NoError(t, db.Where("account_id = ?", created.ID).First(&credential).Error)
	require.NotContains(t, credential.CookieCiphertext, "secret-cookie")
	plainCookie, err := decryptCredential(credential.CookieCiphertext, credentialAAD(created.ID, "cookie"))
	require.NoError(t, err)
	require.Equal(t, "secret-cookie", plainCookie)
	_, err = decryptCredential(credential.CookieCiphertext, credentialAAD(created.ID+1, "cookie"))
	require.Error(t, err)

	ownerList, total, err := service.List(t.Context(), 10, xyRequest.AccountSearch{})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.True(t, ownerList[0].HasCredential)
	otherList, total, err := service.List(t.Context(), 11, xyRequest.AccountSearch{})
	require.NoError(t, err)
	require.Zero(t, total)
	require.Empty(t, otherList)
	require.Error(t, service.Delete(t.Context(), 11, created.ID))
}
