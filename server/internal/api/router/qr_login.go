package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// QRLoginRouter 注册扫码登录 HTTP 路由。
type QRLoginRouter struct {
	handlers *api.API // handlers 提供 HTTP 处理器。
}

// NewQRLoginRouter 创建 QRLoginRouter。
func NewQRLoginRouter(handlers *api.API) *QRLoginRouter {
	return &QRLoginRouter{handlers: handlers}
}

// Init 注册当前模块路由。
func (r *QRLoginRouter) Init(private *gin.RouterGroup) {
	read := private.Group("xy/qr-login")
	write := private.Group("xy/qr-login")
	read.POST("start", r.handlers.QRLogin.Start)
	read.GET("status", r.handlers.QRLogin.Status)
	write.DELETE("cancel", r.handlers.QRLogin.Cancel)
}
