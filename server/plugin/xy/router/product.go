package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type product struct{}

func (r *product) Init(private *gin.RouterGroup) {
	read := private.Group("xy/product")
	write := private.Group("xy/product").Use(middleware.OperationRecord())
	read.GET("list", productApi.List)
	write.POST("sync", productApi.Sync)
	write.PUT("setting", productApi.UpdateSetting)
}
