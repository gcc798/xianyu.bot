package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

type AuthRouter struct{ handlers *api.API }

func NewAuthRouter(handlers *api.API) *AuthRouter { return &AuthRouter{handlers: handlers} }

func (s *AuthRouter) Init(router *gin.RouterGroup) gin.IRoutes {
	base := router.Group("base")
	base.POST("login", s.handlers.User.Login)
	return base
}
