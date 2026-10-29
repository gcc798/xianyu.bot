package service

import (
	"encoding/base64"
	"os"
	"strings"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model"
	xyRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestLiveQRLoginCanGenerateAndCancel(t *testing.T) {
	if os.Getenv("XY_LIVE_TEST") != "1" {
		t.Skip("set XY_LIVE_TEST=1 to call the public Xy login endpoints")
	}
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Credential{}))
	oldDB := global.GVA_DB
	t.Cleanup(func() { global.GVA_DB = oldDB })
	global.GVA_DB = db
	t.Setenv(credentialKeyEnvironment, base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))
	accountService := new(account)
	created, err := accountService.Create(t.Context(), 10, xyRequest.CreateAccount{Label: "在线测试账号"})
	require.NoError(t, err)
	loginService := newQRLogin()

	login, err := loginService.Start(t.Context(), 10, created.ID)
	require.NoError(t, err)
	require.Equal(t, qrStatusPending, login.Status)
	require.True(t, strings.HasPrefix(login.QRCodeDataURL, "data:image/png;base64,"))
	require.NoError(t, loginService.Cancel(10, login.SessionID))
}
