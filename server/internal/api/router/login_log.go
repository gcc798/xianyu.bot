package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gcc798/xianyu.bot/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// LoginLogRouter 注册登录审计 HTTP 路由。
type LoginLogRouter struct {
	handlers *api.API // handlers 提供 HTTP 处理器。
}

// NewLoginLogRouter 创建 LoginLogRouter。
func NewLoginLogRouter(handlers *api.API) *LoginLogRouter {
	return &LoginLogRouter{handlers: handlers}
}

// Init 注册当前模块路由。
func (s *LoginLogRouter) Init(router *gin.RouterGroup) {
	loginLog := router.Group("sysLoginLog").Use(middleware.AdminOnly())
	loginLog.DELETE("deleteLoginLog", s.handlers.LoginLog.Delete)
	loginLog.GET("getLoginLogList", s.handlers.LoginLog.List)
}
