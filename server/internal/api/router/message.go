package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// MessageRouter 注册消息 HTTP 路由。
type MessageRouter struct {
	api *api.API // api 提供 HTTP 处理器。
}

// NewMessageRouter 创建 MessageRouter。
func NewMessageRouter(api *api.API) *MessageRouter { return &MessageRouter{api: api} }

// Init 注册当前模块路由。
func (r *MessageRouter) Init(group *gin.RouterGroup) {
	read := group.Group("xy/message")
	write := group.Group("xy/message")
	read.GET("list", r.api.Message.List)
	write.POST("reply", r.api.Message.Reply)
	write.PUT("takeover", r.api.Message.UpdateTakeover)
}
