package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// WebSocketRouter 注册闲鱼连接 HTTP 路由。
type WebSocketRouter struct {
	api *api.API // api 提供 HTTP 处理器。
}

// NewWebSocketRouter 创建 WebSocketRouter。
func NewWebSocketRouter(api *api.API) *WebSocketRouter {
	return &WebSocketRouter{api: api}
}

// Init 注册当前模块路由。
func (r *WebSocketRouter) Init(private *gin.RouterGroup) {
	read := private.Group("xy/websocket")
	write := private.Group("xy/websocket")
	read.GET("status", r.api.WebSocket.Status)
	write.POST("start", r.api.WebSocket.Start)
	write.POST("stop", r.api.WebSocket.Stop)
}
