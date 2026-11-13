package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// OrderRouter 注册订单 HTTP 路由。
type OrderRouter struct {
	api *api.API // api 提供 HTTP 处理器。
}

// NewOrderRouter 创建 OrderRouter。
func NewOrderRouter(api *api.API) *OrderRouter { return &OrderRouter{api: api} }

// Init 注册当前模块路由。
func (r *OrderRouter) Init(group *gin.RouterGroup) {
	read := group.Group("xy/order")
	write := group.Group("xy/order")
	read.GET("list", r.api.Order.List)
	read.GET("detail", r.api.Order.Detail)
	write.POST("sync-pending", r.api.Order.SyncPending)
}
