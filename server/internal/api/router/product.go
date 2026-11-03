package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// ProductRouter 注册商品 HTTP 路由。
type ProductRouter struct {
	handlers *api.API // handlers 提供 HTTP 处理器。
}

// NewProductRouter 创建 ProductRouter。
func NewProductRouter(handlers *api.API) *ProductRouter {
	return &ProductRouter{handlers: handlers}
}

// Init 注册当前模块路由。
func (r *ProductRouter) Init(private *gin.RouterGroup) {
	read := private.Group("xy/product")
	write := private.Group("xy/product")
	read.GET("list", r.handlers.Product.List)
	write.POST("sync", r.handlers.Product.Sync)
	write.PUT("setting", r.handlers.Product.UpdateSetting)
}
