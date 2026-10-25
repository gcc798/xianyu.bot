package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestTelegramRequestMapsProtectedRoute(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/resources/tasks/7", r.URL.Path)
		require.Equal(t, "page=2", r.URL.RawQuery)
		require.Equal(t, "test-token", r.Header.Get("X-Internal-Token"))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(upstream.Close)
	t.Setenv("TELEGRAM_SERVICE_URL", upstream.URL)
	t.Setenv("TELEGRAM_INTERNAL_TOKEN", "test-token")

	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(http.MethodGet, "/api/xianyu/telegram/tasks/7?page=2", nil)
	response, err := telegramRequest(context)
	require.NoError(t, err)
	t.Cleanup(func() { _ = response.Body.Close() })
	require.Equal(t, http.StatusOK, response.StatusCode)
}
