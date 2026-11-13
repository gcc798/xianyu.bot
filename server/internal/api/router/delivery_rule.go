package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// DeliveryRuleRouter 注册发货规则 HTTP 路由。
type DeliveryRuleRouter struct {
	api *api.API // api 提供 HTTP 处理器。
}

// NewDeliveryRuleRouter 创建 DeliveryRuleRouter。
func NewDeliveryRuleRouter(api *api.API) *DeliveryRuleRouter {
	return &DeliveryRuleRouter{api: api}
}

// Init 注册当前模块路由。
func (r *DeliveryRuleRouter) Init(private *gin.RouterGroup) {
	read := private.Group("xy/delivery-rule")
	write := private.Group("xy/delivery-rule")
	read.GET("list", r.api.DeliveryRule.List)
	read.GET("inventory", r.api.DeliveryRule.Inventory)
	read.GET("task-list", r.api.DeliveryRule.ListTasks)
	write.POST("save", r.api.DeliveryRule.Save)
	write.POST("resource-import", r.api.DeliveryRule.ImportResources)
	write.POST("task-retry", r.api.DeliveryRule.RetryTask)
	write.DELETE("delete", r.api.DeliveryRule.Delete)
}
