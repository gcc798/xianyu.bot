package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

type ProductRouter struct{ handlers *api.API }

func NewProductRouter(handlers *api.API) *ProductRouter {
	return &ProductRouter{handlers: handlers}
}

func (r *ProductRouter) Init(private *gin.RouterGroup) {
	read := private.Group("xy/product")
	write := private.Group("xy/product")
	read.GET("list", r.handlers.Product.List)
	write.POST("sync", r.handlers.Product.Sync)
	write.PUT("setting", r.handlers.Product.UpdateSetting)
}
