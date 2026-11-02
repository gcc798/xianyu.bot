package system

import (
	"github.com/gin-gonic/gin"
)

type JWTRouter struct{ *dependencies }

func (s *JWTRouter) Init(router *gin.RouterGroup) {
	jwtRouter := router.Group("jwt")
	jwtRouter.POST("jsonInBlacklist", s.handlers.JWT.JsonInBlacklist)
}
