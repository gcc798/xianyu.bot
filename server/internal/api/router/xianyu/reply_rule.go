package router

import "github.com/gin-gonic/gin"

type replyRule struct{ *dependencies }

func (r *replyRule) Init(private *gin.RouterGroup) {
	read := private.Group("xy/reply-rule")
	write := private.Group("xy/reply-rule")
	read.GET("list", r.handlers.ReplyRule.List)
	write.POST("save", r.handlers.ReplyRule.Save)
	write.DELETE("delete", r.handlers.ReplyRule.Delete)
}
