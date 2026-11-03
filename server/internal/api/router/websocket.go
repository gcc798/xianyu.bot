package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

type WebSocketRouter struct{ handlers *api.API }

func NewWebSocketRouter(handlers *api.API) *WebSocketRouter {
	return &WebSocketRouter{handlers: handlers}
}

func (r *WebSocketRouter) Init(private *gin.RouterGroup) {
	read := private.Group("xy/websocket")
	write := private.Group("xy/websocket")
	read.GET("status", r.handlers.WebSocket.Status)
	write.POST("start", r.handlers.WebSocket.Start)
	write.POST("stop", r.handlers.WebSocket.Stop)
}
