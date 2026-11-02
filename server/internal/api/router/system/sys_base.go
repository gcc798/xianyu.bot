package system

import (
	"github.com/gin-gonic/gin"
)

type AuthRouter struct{ *dependencies }

func (s *AuthRouter) Init(router *gin.RouterGroup) gin.IRoutes {
	base := router.Group("base")
	base.POST("login", s.handlers.User.Login)
	return base
}
