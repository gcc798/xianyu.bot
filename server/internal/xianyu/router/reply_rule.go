package router

import "github.com/gin-gonic/gin"

type replyRule struct{}

func (r *replyRule) Init(private *gin.RouterGroup) {
	read := private.Group("xy/reply-rule")
	write := private.Group("xy/reply-rule")
	read.GET("list", replyRuleApi.List)
	write.POST("save", replyRuleApi.Save)
	write.DELETE("delete", replyRuleApi.Delete)
}
