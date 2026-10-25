package router

import "github.com/gin-gonic/gin"

type status struct{}

func (r *status) Init(private *gin.RouterGroup) {
	private.Group("xianyu").GET("status", statusApi.GetStatus)
}
