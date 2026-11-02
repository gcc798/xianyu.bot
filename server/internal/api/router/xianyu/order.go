package router

import "github.com/gin-gonic/gin"

type order struct{ *dependencies }

func (r *order) Init(group *gin.RouterGroup) {
	read := group.Group("xy/order")
	write := group.Group("xy/order")
	read.GET("list", r.handlers.Order.List)
	write.POST("sync-pending", r.handlers.Order.SyncPending)
}
