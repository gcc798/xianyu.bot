package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

type QRLoginRouter struct{ handlers *api.API }

func (r *QRLoginRouter) Init(private *gin.RouterGroup) {
	read := private.Group("xy/qr-login")
	write := private.Group("xy/qr-login")
	read.POST("start", r.handlers.QRLogin.Start)
	read.GET("status", r.handlers.QRLogin.Status)
	write.DELETE("cancel", r.handlers.QRLogin.Cancel)
}
