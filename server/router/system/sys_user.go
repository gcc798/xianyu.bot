package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type UserRouter struct{}

func (s *UserRouter) InitUserRouter(router *gin.RouterGroup) {
	user := router.Group("user")
	user.POST("changePassword", baseApi.ChangePassword)
	user.PUT("setSelfInfo", baseApi.SetSelfInfo)
	user.GET("getUserInfo", baseApi.GetUserInfo)

	admin := user.Group("").Use(middleware.AdminOnly())
	admin.POST("admin_register", baseApi.Register)
	admin.POST("getUserList", baseApi.GetUserList)
	admin.DELETE("deleteUser", baseApi.DeleteUser)
	admin.PUT("setUserInfo", baseApi.SetUserInfo)
	admin.POST("resetPassword", baseApi.ResetPassword)
}
