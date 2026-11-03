package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

type OrderRouter struct{ handlers *api.API }

func NewOrderRouter(handlers *api.API) *OrderRouter { return &OrderRouter{handlers: handlers} }

func (r *OrderRouter) Init(group *gin.RouterGroup) {
	read := group.Group("xy/order")
	write := group.Group("xy/order")
	read.GET("list", r.handlers.Order.List)
	write.POST("sync-pending", r.handlers.Order.SyncPending)
}
