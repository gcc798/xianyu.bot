package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type qrLogin struct{}

func (r *qrLogin) Init(private *gin.RouterGroup) {
	read := private.Group("xianyu/qr-login")
	write := private.Group("xianyu/qr-login").Use(middleware.OperationRecord())
	// QR image content must not enter the operation-record response log.
	read.POST("start", qrLoginApi.Start)
	read.GET("status", qrLoginApi.Status)
	write.DELETE("cancel", qrLoginApi.Cancel)
}
