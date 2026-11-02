package router

import "github.com/gin-gonic/gin"

type status struct{ *dependencies }

func (r *status) Init(private *gin.RouterGroup) {
	group := private.Group("xy")
	group.GET("status", r.handlers.Status.GetStatus)
	group.GET("dashboard/overview", r.handlers.Status.GetOverview)
}
