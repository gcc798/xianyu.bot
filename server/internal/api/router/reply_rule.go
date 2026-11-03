package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// ReplyRuleRouter 注册回复规则 HTTP 路由。
type ReplyRuleRouter struct {
	handlers *api.API // handlers 提供 HTTP 处理器。
}

// NewReplyRuleRouter 创建 ReplyRuleRouter。
func NewReplyRuleRouter(handlers *api.API) *ReplyRuleRouter {
	return &ReplyRuleRouter{handlers: handlers}
}

// Init 注册当前模块路由。
func (r *ReplyRuleRouter) Init(private *gin.RouterGroup) {
	read := private.Group("xy/reply-rule")
	write := private.Group("xy/reply-rule")
	read.GET("list", r.handlers.ReplyRule.List)
	write.POST("save", r.handlers.ReplyRule.Save)
	write.DELETE("delete", r.handlers.ReplyRule.Delete)
}
