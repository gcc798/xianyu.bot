package system

import (
	"github.com/gcc798/xianyu.bot/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

type LoginLogRouter struct{ *dependencies }

func (s *LoginLogRouter) Init(router *gin.RouterGroup) {
	loginLog := router.Group("sysLoginLog").Use(middleware.AdminOnly())
	loginLog.DELETE("deleteLoginLog", s.handlers.LoginLog.Delete)
	loginLog.GET("getLoginLogList", s.handlers.LoginLog.List)
}
