package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// StatusRouter 注册状态 HTTP 路由。
type StatusRouter struct {
	handlers *api.API // handlers 提供 HTTP 处理器。
}

// NewStatusRouter 创建 StatusRouter。
func NewStatusRouter(handlers *api.API) *StatusRouter { return &StatusRouter{handlers: handlers} }

// Init 注册当前模块路由。
func (r *StatusRouter) Init(private *gin.RouterGroup) {
	group := private.Group("xy")
	group.GET("status", r.handlers.Status.GetStatus)
	group.GET("dashboard/overview", r.handlers.Status.GetOverview)
}
