package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// AccountRouter 注册闲鱼账号 HTTP 路由。
type AccountRouter struct {
	handlers *api.API // handlers 提供 HTTP 处理器。
}

// NewAccountRouter 创建 AccountRouter。
func NewAccountRouter(handlers *api.API) *AccountRouter {
	return &AccountRouter{handlers: handlers}
}

// Init 注册当前模块路由。
func (r *AccountRouter) Init(private *gin.RouterGroup) {
	read := private.Group("xy/account")
	write := private.Group("xy/account")
	read.GET("list", r.handlers.Account.List)
	write.POST("create", r.handlers.Account.Create)
	write.PUT("update", r.handlers.Account.Update)
	write.DELETE("delete", r.handlers.Account.Delete)
	read.PUT("credential", r.handlers.Account.SaveCredential)
	write.POST("check-login", r.handlers.Account.CheckLogin)
}
