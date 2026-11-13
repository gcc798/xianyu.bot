package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// ReplyRuleRouter 注册回复规则 HTTP 路由。
type ReplyRuleRouter struct {
	api *api.API // api 提供 HTTP 处理器。
}

// NewReplyRuleRouter 创建 ReplyRuleRouter。
func NewReplyRuleRouter(api *api.API) *ReplyRuleRouter {
	return &ReplyRuleRouter{api: api}
}

// Init 注册当前模块路由。
func (r *ReplyRuleRouter) Init(private *gin.RouterGroup) {
	read := private.Group("xy/reply-rule")
	write := private.Group("xy/reply-rule")
	read.GET("list", r.api.ReplyRule.List)
	write.POST("save", r.api.ReplyRule.Save)
	write.DELETE("delete", r.api.ReplyRule.Delete)
}
