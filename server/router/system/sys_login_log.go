package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/api/v1"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type LoginLogRouter struct{}

func (s *LoginLogRouter) InitLoginLogRouter(router *gin.RouterGroup) {
	loginLog := router.Group("sysLoginLog").Use(middleware.AdminOnly())
	api := v1.ApiGroupApp.SystemApiGroup.LoginLogApi
	loginLog.DELETE("deleteLoginLog", api.DeleteLoginLog)
	loginLog.DELETE("deleteLoginLogByIds", api.DeleteLoginLogByIds)
	loginLog.GET("findLoginLog", api.FindLoginLog)
	loginLog.GET("getLoginLogList", api.GetLoginLogList)
}
