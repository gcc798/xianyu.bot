package router

import "github.com/gin-gonic/gin"

type message struct{ *dependencies }

func (r *message) Init(group *gin.RouterGroup) {
	read := group.Group("xy/message")
	write := group.Group("xy/message")
	read.GET("list", r.handlers.Message.List)
	write.POST("reply", r.handlers.Message.Reply)
	write.PUT("takeover", r.handlers.Message.UpdateTakeover)
}
