package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

type ReplyRuleRouter struct{ handlers *api.API }

func NewReplyRuleRouter(handlers *api.API) *ReplyRuleRouter {
	return &ReplyRuleRouter{handlers: handlers}
}

func (r *ReplyRuleRouter) Init(private *gin.RouterGroup) {
	read := private.Group("xy/reply-rule")
	write := private.Group("xy/reply-rule")
	read.GET("list", r.handlers.ReplyRule.List)
	write.POST("save", r.handlers.ReplyRule.Save)
	write.DELETE("delete", r.handlers.ReplyRule.Delete)
}
