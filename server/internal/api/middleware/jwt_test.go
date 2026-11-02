package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	systemReq "github.com/gcc798/xianyu.bot/internal/model/system/request"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAdminOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name        string
		authorityID uint
		wantStatus  int
	}{
		{name: "admin", authorityID: 888, wantStatus: http.StatusNoContent},
		{name: "user", authorityID: 9528, wantStatus: http.StatusForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("claims", &systemReq.CustomClaims{BaseClaims: systemReq.BaseClaims{AuthorityID: test.authorityID}})
			})
			router.Use(AdminOnly())
			router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
			require.Equal(t, test.wantStatus, response.Code)
		})
	}
}
