package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type deliveryRule struct{}

func (r *deliveryRule) Init(private *gin.RouterGroup) {
	read := private.Group("xy/delivery-rule")
	write := private.Group("xy/delivery-rule").Use(middleware.OperationRecord())
	read.GET("list", deliveryRuleApi.List)
	read.GET("inventory", deliveryRuleApi.Inventory)
	read.GET("task-list", deliveryRuleApi.ListTasks)
	write.POST("save", deliveryRuleApi.Save)
	write.POST("resource-import", deliveryRuleApi.ImportResources)
	write.POST("task-retry", deliveryRuleApi.RetryTask)
	write.DELETE("delete", deliveryRuleApi.Delete)
}
