package system

import (
	"github.com/gcc798/xianyu.bot/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

type LoginLogRouter struct{ *dependencies }

func (s *LoginLogRouter) InitLoginLogRouter(router *gin.RouterGroup) {
	loginLog := router.Group("sysLoginLog").Use(middleware.AdminOnly())
	loginLog.DELETE("deleteLoginLog", s.handlers.LoginLog.DeleteLoginLog)
	loginLog.GET("getLoginLogList", s.handlers.LoginLog.GetLoginLogList)
}
