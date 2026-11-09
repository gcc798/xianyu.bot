package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// DeliveryRuleRouter 注册发货规则 HTTP 路由。
type DeliveryRuleRouter struct {
	handlers *api.API // handlers 提供 HTTP 处理器。
}

// NewDeliveryRuleRouter 创建 DeliveryRuleRouter。
func NewDeliveryRuleRouter(handlers *api.API) *DeliveryRuleRouter {
	return &DeliveryRuleRouter{handlers: handlers}
}

// Init 注册当前模块路由。
func (r *DeliveryRuleRouter) Init(private *gin.RouterGroup) {
	read := private.Group("xy/delivery-rule")
	write := private.Group("xy/delivery-rule")
	read.GET("list", r.handlers.DeliveryRule.List)
	read.GET("inventory", r.handlers.DeliveryRule.Inventory)
	read.GET("task-list", r.handlers.DeliveryRule.ListTasks)
	write.POST("save", r.handlers.DeliveryRule.Save)
	write.POST("resource-import", r.handlers.DeliveryRule.ImportResources)
	write.POST("task-retry", r.handlers.DeliveryRule.RetryTask)
	write.DELETE("delete", r.handlers.DeliveryRule.Delete)
}
