package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// MessageRouter 注册消息 HTTP 路由。
type MessageRouter struct {
	handlers *api.API // handlers 提供 HTTP 处理器。
}

// NewMessageRouter 创建 MessageRouter。
func NewMessageRouter(handlers *api.API) *MessageRouter { return &MessageRouter{handlers: handlers} }

// Init 注册当前模块路由。
func (r *MessageRouter) Init(group *gin.RouterGroup) {
	read := group.Group("xy/message")
	write := group.Group("xy/message")
	read.GET("list", r.handlers.Message.List)
	write.POST("reply", r.handlers.Message.Reply)
	write.PUT("takeover", r.handlers.Message.UpdateTakeover)
}
