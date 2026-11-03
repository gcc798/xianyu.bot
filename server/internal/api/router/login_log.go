package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gcc798/xianyu.bot/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

type LoginLogRouter struct{ handlers *api.API }

func NewLoginLogRouter(handlers *api.API) *LoginLogRouter {
	return &LoginLogRouter{handlers: handlers}
}

func (s *LoginLogRouter) Init(router *gin.RouterGroup) {
	loginLog := router.Group("sysLoginLog").Use(middleware.AdminOnly())
	loginLog.DELETE("deleteLoginLog", s.handlers.LoginLog.Delete)
	loginLog.GET("getLoginLogList", s.handlers.LoginLog.List)
}
