package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// JWTRouter 注册令牌 HTTP 路由。
type JWTRouter struct {
	handlers *api.API // handlers 提供 HTTP 处理器。
}

// NewJWTRouter 创建 JWTRouter。
func NewJWTRouter(handlers *api.API) *JWTRouter { return &JWTRouter{handlers: handlers} }

// Init 注册当前模块路由。
func (s *JWTRouter) Init(router *gin.RouterGroup) {
	jwtRouter := router.Group("jwt")
	jwtRouter.POST("jsonInBlacklist", s.handlers.JWT.Revoke)
}
