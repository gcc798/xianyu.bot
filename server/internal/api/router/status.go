package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// StatusRouter 注册状态 HTTP 路由。
type StatusRouter struct {
	api *api.API // api 提供 HTTP 处理器。
}

// NewStatusRouter 创建 StatusRouter。
func NewStatusRouter(api *api.API) *StatusRouter { return &StatusRouter{api: api} }

// Init 注册当前模块路由。
func (r *StatusRouter) Init(private *gin.RouterGroup) {
	group := private.Group("xy")
	group.GET("status", r.api.Status.GetStatus)
	group.GET("dashboard/overview", r.api.Status.GetOverview)
}
