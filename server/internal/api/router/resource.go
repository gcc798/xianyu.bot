package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// ResourceRouter 注册资源 HTTP 路由。
type ResourceRouter struct {
	api *api.API // api 提供 HTTP 处理器。
}

// NewResourceRouter 创建 ResourceRouter。
func NewResourceRouter(api *api.API) *ResourceRouter {
	return &ResourceRouter{api: api}
}

// Init 注册当前模块路由。
func (r *ResourceRouter) Init(private *gin.RouterGroup) {
	private.GET("xy/telegram/resource-options", r.api.Resource.TelegramResourceOptions)
	private.GET("xy/telegram/resource-sources", r.api.Resource.TelegramSources)
	private.POST("xy/telegram/resource-sources/sync", r.api.Resource.SyncTelegramSources)
	private.POST("xy/telegram/resource-tasks", r.api.Resource.CreateTelegramTask)
	private.GET("xy/telegram/resource-tasks", r.api.Resource.ListTelegramTasks)
	private.POST("xy/telegram/resource-tasks/:id/cancel", r.api.Resource.CancelTelegramTask)
	private.GET("xy/resources", r.api.Resource.List)
	private.GET("xy/resources/:id", r.api.Resource.Get)
	private.GET("xy/resources/:id/preview", r.api.Resource.Preview)
	private.GET("xy/resources/:id/download", r.api.Resource.Download)
	private.GET("xy/resource-tasks", r.api.Resource.ListTasks)
	private.GET("xy/resource-tasks/:id", r.api.Resource.GetTask)
	private.POST("xy/resource-tasks/:id/cancel", r.api.Resource.CancelTask)
}

// InitInternal 执行InitInternal操作。
func (r *ResourceRouter) InitInternal(internal *gin.RouterGroup) {
	internal.POST("telegram-resources/import", r.api.Resource.ImportTelegramResource)
	internal.DELETE("telegram-resources/files/:id", r.api.Resource.DeleteTelegramResource)
	internal.POST("resource-tasks/:id/upload", r.api.Resource.PresignDelivery)
	internal.POST("resource-tasks/:id/progress", r.api.Resource.UpdateDelivery)
	internal.POST("resource-tasks/:id/complete", r.api.Resource.CompleteDelivery)
	internal.POST("resource-tasks/:id/fail", r.api.Resource.FailDelivery)
}
