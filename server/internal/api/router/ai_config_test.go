package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	systemReq "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// TestAIConfigAdminOnly 验证 MCP 管理路由拒绝非管理员。
func TestAIConfigAdminOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name        string
		authorityID uint
		wantStatus  int
	}{
		{name: "admin", authorityID: administratorAuthorityID, wantStatus: http.StatusNoContent},
		{name: "user", authorityID: 9528, wantStatus: http.StatusUnauthorized},
	} {
		t.Run(test.name, func(t *testing.T) {
			engine := gin.New()
			engine.Use(func(c *gin.Context) {
				c.Set("claims", &systemReq.CustomClaims{BaseClaims: systemReq.BaseClaims{AuthorityID: test.authorityID}})
			})
			engine.Use(adminOnly())
			engine.POST("/mcp-servers/1/health-check", func(c *gin.Context) { c.Status(http.StatusNoContent) })

			response := httptest.NewRecorder()
			engine.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/mcp-servers/1/health-check", nil))
			require.Equal(t, test.wantStatus, response.Code)
		})
	}
}
