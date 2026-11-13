package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// AuthRouter 注册认证 HTTP 路由。
type AuthRouter struct {
	api *api.API // api 提供 HTTP 处理器。
}

// NewAuthRouter 创建 AuthRouter。
func NewAuthRouter(api *api.API) *AuthRouter { return &AuthRouter{api: api} }

// Init 注册当前模块路由。
func (s *AuthRouter) Init(router *gin.RouterGroup) gin.IRoutes {
	base := router.Group("base")
	base.POST("login", s.api.User.Login)
	return base
}
