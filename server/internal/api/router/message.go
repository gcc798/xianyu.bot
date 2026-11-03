package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

type MessageRouter struct{ handlers *api.API }

func NewMessageRouter(handlers *api.API) *MessageRouter { return &MessageRouter{handlers: handlers} }

func (r *MessageRouter) Init(group *gin.RouterGroup) {
	read := group.Group("xy/message")
	write := group.Group("xy/message")
	read.GET("list", r.handlers.Message.List)
	write.POST("reply", r.handlers.Message.Reply)
	write.PUT("takeover", r.handlers.Message.UpdateTakeover)
}
