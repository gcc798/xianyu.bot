package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type webSocket struct{}

func (r *webSocket) Init(private *gin.RouterGroup) {
	read := private.Group("xianyu/websocket")
	write := private.Group("xianyu/websocket").Use(middleware.OperationRecord())
	read.GET("status", webSocketApi.Status)
	write.POST("start", webSocketApi.Start)
	write.POST("stop", webSocketApi.Stop)
}
