package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

type JWTRouter struct{ handlers *api.API }

func NewJWTRouter(handlers *api.API) *JWTRouter { return &JWTRouter{handlers: handlers} }

func (s *JWTRouter) Init(router *gin.RouterGroup) {
	jwtRouter := router.Group("jwt")
	jwtRouter.POST("jsonInBlacklist", s.handlers.JWT.Revoke)
}
