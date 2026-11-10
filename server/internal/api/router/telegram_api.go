package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// TelegramAPIConfigRouter 注册 Telegram API 配置路由。
type TelegramAPIConfigRouter struct {
	handlers *api.API
}

// NewTelegramAPIConfigRouter 创建 TelegramAPIConfigRouter。
func NewTelegramAPIConfigRouter(handlers *api.API) *TelegramAPIConfigRouter {
	return &TelegramAPIConfigRouter{handlers: handlers}
}

// Init 注册管理员配置接口。
func (r *TelegramAPIConfigRouter) Init(private *gin.RouterGroup) {
	group := private.Group("xy/telegram-api-configs").Use(adminOnly())
	group.GET("", r.handlers.TelegramAPI.List)
	group.POST("", r.handlers.TelegramAPI.Create)
	group.PUT(":id", r.handlers.TelegramAPI.Update)
	group.DELETE(":id", r.handlers.TelegramAPI.Delete)
}

// InitInternal 注册 Telegram 服务内部接口。
func (r *TelegramAPIConfigRouter) InitInternal(internal *gin.RouterGroup) {
	group := internal.Group("telegram-api-configs")
	group.GET("active", r.handlers.TelegramAPI.ActiveCredential)
	group.POST("flood-wait", r.handlers.TelegramAPI.ReportFloodWait)
}
