package router

import "github.com/gin-gonic/gin"

type product struct{ *dependencies }

func (r *product) Init(private *gin.RouterGroup) {
	read := private.Group("xy/product")
	write := private.Group("xy/product")
	read.GET("list", r.handlers.Product.List)
	write.POST("sync", r.handlers.Product.Sync)
	write.PUT("setting", r.handlers.Product.UpdateSetting)
}
