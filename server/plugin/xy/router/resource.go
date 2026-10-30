package router

import "github.com/gin-gonic/gin"

type resource struct{}

func (r *resource) Init(private *gin.RouterGroup) {
	private.GET("xy/resources", resourceAPI.List)
	private.GET("xy/resources/:id", resourceAPI.Get)
	private.GET("xy/resources/:id/download", resourceAPI.Download)
	private.GET("xy/resource-tasks/:id", resourceAPI.GetTask)
}

func (r *resource) InitInternal(internal *gin.RouterGroup) {
	internal.POST("resource-tasks/:id/upload", resourceAPI.PresignDelivery)
	internal.POST("resource-tasks/:id/progress", resourceAPI.UpdateDelivery)
	internal.POST("resource-tasks/:id/complete", resourceAPI.CompleteDelivery)
	internal.POST("resource-tasks/:id/fail", resourceAPI.FailDelivery)
}
