package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// ResourceRouter 注册资源 HTTP 路由。
type ResourceRouter struct {
	handlers *api.API // handlers 提供 HTTP 处理器。
}

// NewResourceRouter 创建 ResourceRouter。
func NewResourceRouter(handlers *api.API) *ResourceRouter {
	return &ResourceRouter{handlers: handlers}
}

// Init 注册当前模块路由。
func (r *ResourceRouter) Init(private *gin.RouterGroup) {
	private.GET("xy/telegram/resource-options", r.handlers.Resource.TelegramResourceOptions)
	private.GET("xy/telegram/resource-sources", r.handlers.Resource.TelegramSources)
	private.POST("xy/telegram/resource-sources/sync", r.handlers.Resource.SyncTelegramSources)
	private.POST("xy/telegram/resource-tasks", r.handlers.Resource.CreateTelegramTask)
	private.GET("xy/telegram/resource-tasks", r.handlers.Resource.ListTelegramTasks)
	private.POST("xy/telegram/resource-tasks/:id/cancel", r.handlers.Resource.CancelTelegramTask)
	private.GET("xy/resources", r.handlers.Resource.List)
	private.GET("xy/resources/:id", r.handlers.Resource.Get)
	private.GET("xy/resources/:id/preview", r.handlers.Resource.Preview)
	private.GET("xy/resources/:id/download", r.handlers.Resource.Download)
	private.GET("xy/resource-tasks", r.handlers.Resource.ListTasks)
	private.GET("xy/resource-tasks/:id", r.handlers.Resource.GetTask)
	private.POST("xy/resource-tasks/:id/cancel", r.handlers.Resource.CancelTask)
}

// InitInternal 执行InitInternal操作。
func (r *ResourceRouter) InitInternal(internal *gin.RouterGroup) {
	internal.POST("telegram-resources/import", r.handlers.Resource.ImportTelegramResource)
	internal.DELETE("telegram-resources/files/:id", r.handlers.Resource.DeleteTelegramResource)
	internal.POST("resource-tasks/:id/upload", r.handlers.Resource.PresignDelivery)
	internal.POST("resource-tasks/:id/progress", r.handlers.Resource.UpdateDelivery)
	internal.POST("resource-tasks/:id/complete", r.handlers.Resource.CompleteDelivery)
	internal.POST("resource-tasks/:id/fail", r.handlers.Resource.FailDelivery)
}
