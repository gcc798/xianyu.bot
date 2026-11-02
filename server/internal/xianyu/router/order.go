package router

import "github.com/gin-gonic/gin"

type order struct{}

func (r *order) Init(group *gin.RouterGroup) {
	read := group.Group("xy/order")
	write := group.Group("xy/order")
	read.GET("list", orderApi.List)
	write.POST("sync-pending", orderApi.SyncPending)
}
