package service

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gcc798/go.xianyu.assistant/server/internal/infra"
	"github.com/gcc798/go.xianyu.assistant/server/internal/model/xianyu"
	xyRequest "github.com/gcc798/go.xianyu.assistant/server/internal/model/xianyu/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestQRLoginSavesCredentialWithoutReturningIt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/h5":
			http.SetCookie(writer, &http.Cookie{Name: "_m_h5_tk", Value: "token_expiry", Path: "/"})
			_, _ = writer.Write([]byte(`{"ret":["SUCCESS::调用成功"]}`))
		case "/mini":
			_, _ = writer.Write([]byte(`<script>window.viewData = {"loginFormData":{"appName":"xianyu","fromSite":"77"}};</script>`))
		case "/generate":
			if request.URL.Query().Get("appName") != "xianyu" {
				t.Errorf("missing generated login parameters")
			}
			_, _ = writer.Write([]byte(`{"content":{"success":true,"data":{"t":"time-demo","ck":"check-demo","codeContent":"xianyu-login-demo"}}}`))
		case "/query":
			http.SetCookie(writer, &http.Cookie{Name: "unb", Value: "user-demo", Path: "/"})
			http.SetCookie(writer, &http.Cookie{Name: "cookie2", Value: "secret-demo", Path: "/"})
			_, _ = writer.Write([]byte(`{"content":{"data":{"qrCodeStatus":"CONFIRMED","token":"login-demo","iframeRedirect":false}}}`))
		case "/login-token":
			require.Equal(t, "login-demo", request.URL.Query().Get("token"))
			require.NoError(t, request.ParseForm())
			http.SetCookie(writer, &http.Cookie{Name: "sgcookie", Value: "login-secret", Path: "/"})
			_, _ = writer.Write([]byte(`{"content":{"success":true}}`))
		default:
			http.NotFound(writer, request)
		}
	}))
	t.Cleanup(server.Close)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Credential{}))
	oldDB := infra.DB
	t.Cleanup(func() { infra.DB = oldDB })
	infra.DB = db
	t.Setenv(credentialKeyEnvironment, base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))
	accountService := &account{db: infra.DB}
	created, err := accountService.Create(t.Context(), 10, xyRequest.CreateAccount{Label: "扫码账号"})
	require.NoError(t, err)

	loginService := &qrLogin{manager: &qrLoginManager{
		accounts: accountService,
		sessions: make(map[string]*qrSession), byAccount: make(map[[2]uint]string),
		endpoints: qrEndpoints{
			h5Token: server.URL + "/h5", miniLogin: server.URL + "/mini",
			generate: server.URL + "/generate", query: server.URL + "/query", loginToken: server.URL + "/login-token",
		},
		timeout: time.Second, pollEvery: time.Millisecond,
	}}
	started, err := loginService.Start(t.Context(), 10, created.ID)
	require.NoError(t, err)
	require.Equal(t, qrStatusPending, started.Status)
	require.True(t, strings.HasPrefix(started.QRCodeDataURL, "data:image/png;base64,"))
	require.NotContains(t, started.QRCodeDataURL, "secret-demo")

	require.Eventually(t, func() bool {
		status, statusErr := loginService.Status(10, started.SessionID)
		return statusErr == nil && status.Status == qrStatusConfirmed
	}, time.Second, time.Millisecond)
	status, err := loginService.Status(10, started.SessionID)
	require.NoError(t, err)
	require.Empty(t, status.QRCodeDataURL)
	_, err = loginService.Status(11, started.SessionID)
	require.Error(t, err)

	var credential model.Credential
	require.NoError(t, db.Where("account_id = ?", created.ID).First(&credential).Error)
	plain, err := decryptCredential(credential.CookieCiphertext, credentialAAD(created.ID, "cookie"))
	require.NoError(t, err)
	require.Contains(t, plain, "unb=user-demo")
	require.Contains(t, plain, "cookie2=secret-demo")
	require.Contains(t, plain, "sgcookie=login-secret")
	var updated model.Account
	require.NoError(t, db.First(&updated, created.ID).Error)
	require.Equal(t, "user-demo", updated.ExternalUserID)
}

func TestSafeVerificationURLOnlyAllowsGoofishHTTPS(t *testing.T) {
	require.Equal(t, "https://passport.goofish.com/verify?id=demo", safeVerificationURL("https://passport.goofish.com/verify?id=demo"))
	require.Empty(t, safeVerificationURL("http://passport.goofish.com/verify"))
	require.Empty(t, safeVerificationURL("https://goofish.com.example.com/verify"))
	require.Empty(t, safeVerificationURL("javascript:alert(1)"))
}

func TestStringOrNumberPreservesLoginParameter(t *testing.T) {
	for input, expected := range map[string]string{
		`"time-demo"`:   "time-demo",
		`1723456789012`: "1723456789012",
	} {
		var value stringOrNumber
		require.NoError(t, value.UnmarshalJSON([]byte(input)))
		require.Equal(t, expected, string(value))
	}
}
