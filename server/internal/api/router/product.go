package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// ProductRouter 注册商品 HTTP 路由。
type ProductRouter struct {
	api *api.API // api 提供 HTTP 处理器。
}

// NewProductRouter 创建 ProductRouter。
func NewProductRouter(api *api.API) *ProductRouter {
	return &ProductRouter{api: api}
}

// Init 注册当前模块路由。
func (r *ProductRouter) Init(private *gin.RouterGroup) {
	read := private.Group("xy/product")
	write := private.Group("xy/product")
	read.GET("list", r.api.Product.List)
	read.GET("detail", r.api.Product.Detail)
	write.POST("sync", r.api.Product.Sync)
	write.PUT("setting", r.api.Product.UpdateSetting)
}
