package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

type StatusRouter struct{ handlers *api.API }

func (r *StatusRouter) Init(private *gin.RouterGroup) {
	group := private.Group("xy")
	group.GET("status", r.handlers.Status.GetStatus)
	group.GET("dashboard/overview", r.handlers.Status.GetOverview)
}
