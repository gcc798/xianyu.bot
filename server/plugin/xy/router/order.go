package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type order struct{}

func (r *order) Init(group *gin.RouterGroup) {
	read := group.Group("xy/order")
	write := group.Group("xy/order").Use(middleware.OperationRecord())
	read.GET("list", orderApi.List)
	write.POST("sync-pending", orderApi.SyncPending)
}
