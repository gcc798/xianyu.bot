package system

import (
	"github.com/gin-gonic/gin"
)

type JwtRouter struct{ *dependencies }

func (s *JwtRouter) InitJwtRouter(Router *gin.RouterGroup) {
	jwtRouter := Router.Group("jwt")
	{
		jwtRouter.POST("jsonInBlacklist", s.handlers.JWT.JsonInBlacklist) // jwt加入黑名单
	}
}
