package router

import "github.com/gin-gonic/gin"

type qrLogin struct{}

func (r *qrLogin) Init(private *gin.RouterGroup) {
	read := private.Group("xy/qr-login")
	write := private.Group("xy/qr-login")
	read.POST("start", qrLoginApi.Start)
	read.GET("status", qrLoginApi.Status)
	write.DELETE("cancel", qrLoginApi.Cancel)
}
