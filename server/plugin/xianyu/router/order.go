package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type order struct{}

func (r *order) Init(group *gin.RouterGroup) {
	read := group.Group("xianyu/order")
	write := group.Group("xianyu/order").Use(middleware.OperationRecord())
	read.GET("list", orderApi.List)
	write.POST("sync-pending", orderApi.SyncPending)
}
