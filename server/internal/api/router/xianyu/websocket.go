package router

import "github.com/gin-gonic/gin"

type webSocket struct{ *dependencies }

func (r *webSocket) Init(private *gin.RouterGroup) {
	read := private.Group("xy/websocket")
	write := private.Group("xy/websocket")
	read.GET("status", r.handlers.WebSocket.Status)
	write.POST("start", r.handlers.WebSocket.Start)
	write.POST("stop", r.handlers.WebSocket.Stop)
}
