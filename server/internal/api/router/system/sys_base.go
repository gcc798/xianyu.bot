package system

import (
	"github.com/gin-gonic/gin"
)

type BaseRouter struct{ *dependencies }

func (s *BaseRouter) InitBaseRouter(Router *gin.RouterGroup) (R gin.IRoutes) {
	baseRouter := Router.Group("base")
	{
		baseRouter.POST("login", s.handlers.Base.Login)
	}
	return baseRouter
}
