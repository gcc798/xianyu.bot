package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// TestTelegramRequestPreservesUpstreamPath 验证对应测试场景。
func TestTelegramRequestPreservesUpstreamPath(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/resources/tasks/7", r.URL.Path)
		require.Equal(t, "page=2", r.URL.RawQuery)
		require.Empty(t, r.Header.Get("X-Internal-Token"))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(upstream.Close)
	t.Setenv("TELEGRAM_SERVICE_URL", upstream.URL)

	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Params = gin.Params{{Key: "path", Value: "/api/resources/tasks/7"}}
	context.Request = httptest.NewRequest(http.MethodGet, "/api/xy/proxy/telegram/api/resources/tasks/7?page=2", nil)
	result, err := telegramRequest(context)
	require.NoError(t, err)
	t.Cleanup(func() { _ = result.Body.Close() })
	require.Equal(t, http.StatusOK, result.StatusCode)
}

// TestTelegramRequestDoesNotForwardSystemCredentials 验证对应测试场景。
func TestTelegramRequestDoesNotForwardSystemCredentials(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Empty(t, r.Header.Get("X-Token"))
		require.Empty(t, r.Header.Get("X-User-Id"))
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(upstream.Close)
	t.Setenv("TELEGRAM_SERVICE_URL", upstream.URL)

	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Params = gin.Params{{Key: "path", Value: "/healthz"}}
	context.Request = httptest.NewRequest(http.MethodGet, "/api/xy/proxy/telegram/healthz", nil)
	context.Request.Header.Set("X-Token", "session-token")
	context.Request.Header.Set("X-User-Id", "1")
	result, err := telegramRequest(context)
	require.NoError(t, err)
	t.Cleanup(func() { _ = result.Body.Close() })
}
