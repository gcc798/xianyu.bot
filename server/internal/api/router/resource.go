package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

type ResourceRouter struct{ handlers *api.API }

func (r *ResourceRouter) Init(private *gin.RouterGroup) {
	private.GET("xy/resources", r.handlers.Resource.List)
	private.GET("xy/resources/:id", r.handlers.Resource.Get)
	private.GET("xy/resources/:id/download", r.handlers.Resource.Download)
	private.GET("xy/resource-tasks/:id", r.handlers.Resource.GetTask)
	private.POST("xy/resource-tasks/:id/cancel", r.handlers.Resource.CancelTask)
}

func (r *ResourceRouter) InitInternal(internal *gin.RouterGroup) {
	internal.POST("resource-tasks/:id/upload", r.handlers.Resource.PresignDelivery)
	internal.POST("resource-tasks/:id/progress", r.handlers.Resource.UpdateDelivery)
	internal.POST("resource-tasks/:id/complete", r.handlers.Resource.CompleteDelivery)
	internal.POST("resource-tasks/:id/fail", r.handlers.Resource.FailDelivery)
}
