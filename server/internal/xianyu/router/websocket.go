package router

import "github.com/gin-gonic/gin"

type webSocket struct{}

func (r *webSocket) Init(private *gin.RouterGroup) {
	read := private.Group("xy/websocket")
	write := private.Group("xy/websocket")
	read.GET("status", webSocketApi.Status)
	write.POST("start", webSocketApi.Start)
	write.POST("stop", webSocketApi.Stop)
}
