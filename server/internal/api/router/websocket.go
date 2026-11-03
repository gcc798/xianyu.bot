package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// WebSocketRouter 注册闲鱼连接 HTTP 路由。
type WebSocketRouter struct {
	handlers *api.API // handlers 提供 HTTP 处理器。
}

// NewWebSocketRouter 创建 WebSocketRouter。
func NewWebSocketRouter(handlers *api.API) *WebSocketRouter {
	return &WebSocketRouter{handlers: handlers}
}

// Init 注册当前模块路由。
func (r *WebSocketRouter) Init(private *gin.RouterGroup) {
	read := private.Group("xy/websocket")
	write := private.Group("xy/websocket")
	read.GET("status", r.handlers.WebSocket.Status)
	write.POST("start", r.handlers.WebSocket.Start)
	write.POST("stop", r.handlers.WebSocket.Stop)
}
