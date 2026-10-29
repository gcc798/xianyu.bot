package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type message struct{}

func (r *message) Init(group *gin.RouterGroup) {
	read := group.Group("xy/message")
	write := group.Group("xy/message").Use(middleware.OperationRecord())
	read.GET("list", messageApi.List)
	write.POST("reply", messageApi.Reply)
	write.PUT("takeover", messageApi.UpdateTakeover)
}
