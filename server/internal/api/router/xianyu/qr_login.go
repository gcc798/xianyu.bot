package router

import "github.com/gin-gonic/gin"

type qrLogin struct{ *dependencies }

func (r *qrLogin) Init(private *gin.RouterGroup) {
	read := private.Group("xy/qr-login")
	write := private.Group("xy/qr-login")
	read.POST("start", r.handlers.QRLogin.Start)
	read.GET("status", r.handlers.QRLogin.Status)
	write.DELETE("cancel", r.handlers.QRLogin.Cancel)
}
