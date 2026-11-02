package router

import "github.com/gin-gonic/gin"

type resource struct{ *dependencies }

func (r *resource) Init(private *gin.RouterGroup) {
	private.GET("xy/resources", r.handlers.Resource.List)
	private.GET("xy/resources/:id", r.handlers.Resource.Get)
	private.GET("xy/resources/:id/download", r.handlers.Resource.Download)
	private.GET("xy/resource-tasks/:id", r.handlers.Resource.GetTask)
	private.POST("xy/resource-tasks/:id/cancel", r.handlers.Resource.CancelTask)
}

func (r *resource) InitInternal(internal *gin.RouterGroup) {
	internal.POST("resource-tasks/:id/upload", r.handlers.Resource.PresignDelivery)
	internal.POST("resource-tasks/:id/progress", r.handlers.Resource.UpdateDelivery)
	internal.POST("resource-tasks/:id/complete", r.handlers.Resource.CompleteDelivery)
	internal.POST("resource-tasks/:id/fail", r.handlers.Resource.FailDelivery)
}
