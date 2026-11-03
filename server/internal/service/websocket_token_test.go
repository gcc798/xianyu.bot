package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestWebSocketTokenIsEncryptedAndCached(t *testing.T) {
	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/device" {
			http.SetCookie(writer, &http.Cookie{Name: "cna", Value: "device-cookie", Path: "/"})
			return
		}
		calls.Add(1)
		require.NoError(t, request.ParseForm())
		var data map[string]string
		require.NoError(t, json.Unmarshal([]byte(request.Form.Get("data")), &data))
		require.NotEmpty(t, data["deviceId"])
		require.NotEmpty(t, request.URL.Query().Get("sign"))
		require.Equal(t, "a21ybx.im.0.0", request.URL.Query().Get("spm_cnt"))
		require.Equal(t, "same-site", request.Header.Get("Sec-Fetch-Site"))
		require.Contains(t, request.Header.Get("Cookie"), "cna=device-cookie")
		http.SetCookie(writer, &http.Cookie{Name: "_m_h5_tk", Value: "new-token_expiry", Path: "/"})
		_, _ = writer.Write([]byte(`{"ret":["SUCCESS::调用成功"],"data":{"accessToken":"secret-websocket-token"}}`))
	}))
	t.Cleanup(server.Close)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Credential{}))
	t.Setenv(credentialKeyEnvironment, base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))
	accountService := &Account{db: db}
	created, err := accountService.Create(t.Context(), 10, xyRequest.CreateAccount{Label: "Token 账号"})
	require.NoError(t, err)
	require.NoError(t, accountService.SaveCredential(t.Context(), 10, xyRequest.SaveCredential{
		ID: created.ID, Cookie: "unb=user-demo; _m_h5_tk=old-token_expiry; cookie2=session",
	}))

	tokenService := NewWebSocketToken(db, &BusinessEvent{db: db})
	tokenService.endpoint = server.URL
	tokenService.deviceCookieEndpoint = server.URL + "/device"
	token, deviceID, err := tokenService.Get(t.Context(), 10, created.ID)
	require.NoError(t, err)
	require.Equal(t, "secret-websocket-token", token)
	require.Contains(t, deviceID, "-user-demo")

	var credential model.Credential
	require.NoError(t, db.Where("account_id = ?", created.ID).First(&credential).Error)
	require.NotContains(t, credential.TokenCiphertext, token)
	require.Equal(t, uint(2), credential.Version)
	require.WithinDuration(t, time.Now().Add(webSocketTokenLifetime), *credential.TokenExpiresAt, time.Minute)
	plainToken, err := decryptCredential(db, credential.TokenCiphertext, credentialAAD(created.ID, "token"))
	require.NoError(t, err)
	require.Equal(t, token, plainToken)
	plainCookie, err := decryptCredential(db, credential.CookieCiphertext, credentialAAD(created.ID, "cookie"))
	require.NoError(t, err)
	require.Contains(t, plainCookie, "_m_h5_tk=new-token_expiry")

	cached, cachedDeviceID, err := tokenService.Get(t.Context(), 10, created.ID)
	require.NoError(t, err)
	require.Equal(t, token, cached)
	require.Equal(t, deviceID, cachedDeviceID)
	require.Equal(t, int64(1), calls.Load())
	_, _, err = tokenService.Get(t.Context(), 11, created.ID)
	require.Error(t, err)
}

func TestWebSocketTokenMarksAccountWhenVerificationIsRequired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		http.SetCookie(writer, &http.Cookie{Name: "x5sec", Value: "challenge-cookie", Path: "/"})
		_, _ = writer.Write([]byte(`{"ret":["FAIL_SYS_USER_VALIDATE::哎哟喂,被挤爆啦,请稍后重试"],"data":{"url":"https://passport.goofish.com/verify"}}`))
	}))
	t.Cleanup(server.Close)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Credential{}, &model.BusinessEvent{}))
	t.Setenv(credentialKeyEnvironment, base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))
	accountService := &Account{db: db}
	created, err := accountService.Create(t.Context(), 10, xyRequest.CreateAccount{Label: "需验证账号"})
	require.NoError(t, err)
	require.NoError(t, accountService.SaveCredential(t.Context(), 10, xyRequest.SaveCredential{
		ID: created.ID, Cookie: "unb=user-demo; cna=device-cookie; _m_h5_tk=old-token_expiry; cookie2=session",
	}))

	tokenService := NewWebSocketToken(db, &BusinessEvent{db: db})
	tokenService.endpoint = server.URL
	_, _, err = tokenService.Get(t.Context(), 10, created.ID)
	verification, ok := errors.AsType[*VerificationRequiredError](err)
	require.True(t, ok)
	require.True(t, verification.NonRetryable())
	require.NoError(t, db.First(&created, created.ID).Error)
	require.Equal(t, "verification_required", created.Status)
	require.NotEmpty(t, created.DeviceID)
	var credential model.Credential
	require.NoError(t, db.Where("account_id = ?", created.ID).First(&credential).Error)
	cookie, err := decryptCredential(db, credential.CookieCiphertext, credentialAAD(created.ID, "cookie"))
	require.NoError(t, err)
	require.Contains(t, cookie, "x5sec=challenge-cookie")
}
