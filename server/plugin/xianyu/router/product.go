package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type product struct{}

func (r *product) Init(private *gin.RouterGroup) {
	read := private.Group("xianyu/product")
	write := private.Group("xianyu/product").Use(middleware.OperationRecord())
	read.GET("list", productApi.List)
	write.POST("sync", productApi.Sync)
	write.PUT("setting", productApi.UpdateSetting)
}
