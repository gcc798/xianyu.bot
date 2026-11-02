package system

import (
	"github.com/gcc798/xianyu.bot/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

type UserRouter struct{ *dependencies }

func (s *UserRouter) Init(router *gin.RouterGroup) {
	user := router.Group("user")
	user.POST("changePassword", s.handlers.User.ChangePassword)
	user.PUT("setSelfInfo", s.handlers.User.SetSelfInfo)
	user.GET("getUserInfo", s.handlers.User.GetUserInfo)

	admin := user.Group("").Use(middleware.AdminOnly())
	admin.POST("admin_register", s.handlers.User.Register)
	admin.POST("getUserList", s.handlers.User.GetUserList)
	admin.DELETE("deleteUser", s.handlers.User.DeleteUser)
	admin.PUT("setUserInfo", s.handlers.User.SetUserInfo)
	admin.POST("resetPassword", s.handlers.User.ResetPassword)
}
