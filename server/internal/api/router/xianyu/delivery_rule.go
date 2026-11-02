package router

import "github.com/gin-gonic/gin"

type deliveryRule struct{ *dependencies }

func (r *deliveryRule) Init(private *gin.RouterGroup) {
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
