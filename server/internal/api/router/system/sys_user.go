package system

import (
	"github.com/gcc798/xianyu.bot/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

type UserRouter struct{ *dependencies }

func (s *UserRouter) InitUserRouter(router *gin.RouterGroup) {
	user := router.Group("user")
	user.POST("changePassword", s.handlers.Base.ChangePassword)
	user.PUT("setSelfInfo", s.handlers.Base.SetSelfInfo)
	user.GET("getUserInfo", s.handlers.Base.GetUserInfo)

	admin := user.Group("").Use(middleware.AdminOnly())
	admin.POST("admin_register", s.handlers.Base.Register)
	admin.POST("getUserList", s.handlers.Base.GetUserList)
	admin.DELETE("deleteUser", s.handlers.Base.DeleteUser)
	admin.PUT("setUserInfo", s.handlers.Base.SetUserInfo)
	admin.POST("resetPassword", s.handlers.Base.ResetPassword)
}
