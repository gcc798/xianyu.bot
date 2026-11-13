package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// QRLoginRouter 注册扫码登录 HTTP 路由。
type QRLoginRouter struct {
	api *api.API // api 提供 HTTP 处理器。
}

// NewQRLoginRouter 创建 QRLoginRouter。
func NewQRLoginRouter(api *api.API) *QRLoginRouter {
	return &QRLoginRouter{api: api}
}

// Init 注册当前模块路由。
func (r *QRLoginRouter) Init(private *gin.RouterGroup) {
	read := private.Group("xy/qr-login")
	write := private.Group("xy/qr-login")
	read.POST("start", r.api.QRLogin.Start)
	read.GET("status", r.api.QRLogin.Status)
	write.DELETE("cancel", r.api.QRLogin.Cancel)
}
