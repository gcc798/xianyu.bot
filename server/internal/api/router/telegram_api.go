package router

import (
	"crypto/subtle"
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strings"
)

// TelegramAPIConfigRouter 注册 Telegram API 配置路由。
type TelegramAPIConfigRouter struct {
	api *api.API // api 提供 HTTP 处理器。
}

// NewTelegramAPIConfigRouter 创建 TelegramAPIConfigRouter。
func NewTelegramAPIConfigRouter(api *api.API) *TelegramAPIConfigRouter {
	return &TelegramAPIConfigRouter{api: api}
}

// Init 注册管理员配置接口。
func (r *TelegramAPIConfigRouter) Init(private *gin.RouterGroup) {
	group := private.Group("xy/telegram-api-configs").Use(adminOnly())
	group.GET("", r.api.TelegramAPI.List)
	group.POST("", r.api.TelegramAPI.Create)
	group.PUT(":id", r.api.TelegramAPI.Update)
	group.DELETE(":id", r.api.TelegramAPI.Delete)
}

// InitInternal 注册 Telegram 服务内部接口。
func (r *TelegramAPIConfigRouter) InitInternal(internal *gin.RouterGroup) {
	token, _ := os.ReadFile("/run/xy/token")
	secret := strings.TrimSpace(string(token))
	group := internal.Group("telegram-api-configs", func(c *gin.Context) {
		if secret == "" || subtle.ConstantTimeCompare([]byte(secret), []byte(c.GetHeader("X-Service-Token"))) != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 7, "data": nil, "msg": "服务身份无效"})
		}
	})
	group.GET("active", r.api.TelegramAPI.ActiveCredential)
	group.POST("flood-wait", r.api.TelegramAPI.ReportFloodWait)
}
