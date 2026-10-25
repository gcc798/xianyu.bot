package router

import "github.com/gin-gonic/gin"

type telegramResource struct{}

func (r *telegramResource) Init(private *gin.RouterGroup) {
	group := private.Group("xianyu/telegram")
	group.GET("options", telegramResourceApi.ProxyJSON)
	group.POST("tasks", telegramResourceApi.ProxyJSON)
	group.GET("tasks", telegramResourceApi.ProxyJSON)
	group.GET("tasks/:task_id", telegramResourceApi.ProxyJSON)
	group.POST("tasks/:task_id/cancel", telegramResourceApi.ProxyJSON)
	group.POST("tasks/:task_id/retry", telegramResourceApi.ProxyJSON)
	group.GET("files", telegramResourceApi.ProxyJSON)
	group.GET("files/:file_id/preview", telegramResourceApi.ProxyFile)
	group.GET("files/:file_id/download", telegramResourceApi.ProxyFile)
}
