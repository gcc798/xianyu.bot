package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

type DeliveryRuleRouter struct{ handlers *api.API }

func NewDeliveryRuleRouter(handlers *api.API) *DeliveryRuleRouter {
	return &DeliveryRuleRouter{handlers: handlers}
}

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
