package service

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestHTTPLoginCheckerMergesResponseCookies 验证对应测试场景。
func TestHTTPLoginCheckerMergesResponseCookies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || !strings.Contains(request.Header.Get("Cookie"), "unb=user-demo") {
			t.Errorf("unexpected request method or cookie")
		}
		if err := request.ParseForm(); err != nil || request.Form.Get("hid") != "user-demo" {
			t.Errorf("unexpected login form: %v", err)
		}
		http.SetCookie(writer, &http.Cookie{Name: "_m_h5_tk", Value: "new-token"})
		_, _ = writer.Write([]byte(`{"content":{"success":true}}`))
	}))
	t.Cleanup(server.Close)
	checker := &HTTPLoginChecker{client: server.Client(), endpoint: server.URL}

	result, err := checker.Check(t.Context(), "unb=user-demo; _m_h5_tk=old-token; cookie2=value=with=equals")
	require.NoError(t, err)
	require.True(t, result.valid)
	require.False(t, result.risk)
	require.Contains(t, result.cookie, "_m_h5_tk=new-token")
	require.Contains(t, result.cookie, "cookie2=value=with=equals")
}

// TestCheckLoginSerializesRefreshAndUpdatesEncryptedVersion 验证对应测试场景。
func TestCheckLoginSerializesRefreshAndUpdatesEncryptedVersion(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Credential{}))
	t.Setenv(credentialKeyEnvironment, base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))

	checker := newSerializedLoginChecker()
	service := &Account{db: db, events: &BusinessEvent{db: db}, loginChecker: checker}
	created, err := service.Create(t.Context(), 10, xyRequest.CreateAccount{Label: "扫码账号"})
	require.NoError(t, err)
	require.NoError(t, service.SaveCredential(t.Context(), 10, xyRequest.SaveCredential{ID: created.ID, Cookie: "unb=old; token=old"}))

	errorsCh := make(chan error, 2)
	for range 2 {
		go func() {
			_, checkErr := service.CheckLogin(t.Context(), 10, created.ID)
			errorsCh <- checkErr
		}()
	}
	require.NoError(t, <-errorsCh)
	require.NoError(t, <-errorsCh)
	require.Equal(t, int64(1), checker.maxActive.Load())
	_, err = service.CheckLogin(t.Context(), 11, created.ID)
	require.Error(t, err)

	var credential model.Credential
	require.NoError(t, db.Where("account_id = ?", created.ID).First(&credential).Error)
	require.Equal(t, uint(2), credential.Version)
	plain, err := decryptCredential(db, credential.CookieCiphertext, credentialAAD(created.ID, "cookie"))
	require.NoError(t, err)
	require.Equal(t, "unb=updated; token=new", plain)
	var updated model.Account
	require.NoError(t, db.First(&updated, created.ID).Error)
	require.Equal(t, "ready", updated.Status)
}

// serializedLoginChecker 是测试辅助类型。
type serializedLoginChecker struct {
	active    atomic.Int64 // active 记录当前并发数。
	maxActive atomic.Int64 // maxActive 记录最大并发数。
}

// newSerializedLoginChecker 提供测试辅助能力。
func newSerializedLoginChecker() *serializedLoginChecker { return new(serializedLoginChecker) }

// Check 提供测试辅助能力。
func (c *serializedLoginChecker) Check(ctx context.Context, cookie string) (LoginCheck, error) {
	active := c.active.Add(1)
	defer c.active.Add(-1)
	for {
		maximum := c.maxActive.Load()
		if active <= maximum || c.maxActive.CompareAndSwap(maximum, active) {
			break
		}
	}
	select {
	case <-ctx.Done():
		return LoginCheck{}, context.Cause(ctx)
	case <-time.After(10 * time.Millisecond):
	}
	if strings.Contains(cookie, "updated") {
		return LoginCheck{valid: true, cookie: cookie}, nil
	}
	return LoginCheck{valid: true, cookie: "unb=updated; token=new"}, nil
}
