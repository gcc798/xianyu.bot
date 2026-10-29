package router

import "github.com/gin-gonic/gin"

type status struct{}

func (r *status) Init(private *gin.RouterGroup) {
	group := private.Group("xianyu")
	group.GET("status", statusApi.GetStatus)
	group.GET("dashboard/overview", statusApi.GetOverview)
}
