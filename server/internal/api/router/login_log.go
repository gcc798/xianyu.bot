package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gcc798/xianyu.bot/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// LoginLogRouter 注册登录审计 HTTP 路由。
type LoginLogRouter struct {
	api *api.API // api 提供 HTTP 处理器。
}

// NewLoginLogRouter 创建 LoginLogRouter。
func NewLoginLogRouter(api *api.API) *LoginLogRouter {
	return &LoginLogRouter{api: api}
}

// Init 注册当前模块路由。
func (s *LoginLogRouter) Init(router *gin.RouterGroup) {
	loginLog := router.Group("sysLoginLog").Use(middleware.AdminOnly())
	loginLog.DELETE("deleteLoginLog", s.api.LoginLog.Delete)
	loginLog.GET("getLoginLogList", s.api.LoginLog.List)
}
