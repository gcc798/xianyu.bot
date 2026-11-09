package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// OrderRouter 注册订单 HTTP 路由。
type OrderRouter struct {
	handlers *api.API // handlers 提供 HTTP 处理器。
}

// NewOrderRouter 创建 OrderRouter。
func NewOrderRouter(handlers *api.API) *OrderRouter { return &OrderRouter{handlers: handlers} }

// Init 注册当前模块路由。
func (r *OrderRouter) Init(group *gin.RouterGroup) {
	read := group.Group("xy/order")
	write := group.Group("xy/order")
	read.GET("list", r.handlers.Order.List)
	read.GET("detail", r.handlers.Order.Detail)
	write.POST("sync-pending", r.handlers.Order.SyncPending)
}
