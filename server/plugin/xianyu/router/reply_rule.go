package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type replyRule struct{}

func (r *replyRule) Init(private *gin.RouterGroup) {
	read := private.Group("xianyu/reply-rule")
	write := private.Group("xianyu/reply-rule").Use(middleware.OperationRecord())
	read.GET("list", replyRuleApi.List)
	write.POST("save", replyRuleApi.Save)
	write.DELETE("delete", replyRuleApi.Delete)
}
