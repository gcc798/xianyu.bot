package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gcc798/xianyu.bot/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// UserRouter 注册用户 HTTP 路由。
type UserRouter struct {
	api *api.API // api 提供 HTTP 处理器。
}

// NewUserRouter 创建 UserRouter。
func NewUserRouter(api *api.API) *UserRouter { return &UserRouter{api: api} }

// Init 注册当前模块路由。
func (s *UserRouter) Init(router *gin.RouterGroup) {
	user := router.Group("user")
	user.POST("changePassword", s.api.User.ChangePassword)
	user.PUT("setSelfInfo", s.api.User.SetSelfInfo)
	user.GET("getUserInfo", s.api.User.GetUserInfo)

	admin := user.Group("").Use(middleware.AdminOnly())
	admin.POST("admin_register", s.api.User.Register)
	admin.POST("getUserList", s.api.User.GetUserList)
	admin.DELETE("deleteUser", s.api.User.DeleteUser)
	admin.PUT("setUserInfo", s.api.User.SetUserInfo)
	admin.POST("resetPassword", s.api.User.ResetPassword)
}
