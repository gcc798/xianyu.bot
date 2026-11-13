package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// AccountRouter 注册闲鱼账号 HTTP 路由。
type AccountRouter struct {
	api *api.API // api 提供 HTTP 处理器。
}

// NewAccountRouter 创建 AccountRouter。
func NewAccountRouter(api *api.API) *AccountRouter {
	return &AccountRouter{api: api}
}

// Init 注册当前模块路由。
func (r *AccountRouter) Init(private *gin.RouterGroup) {
	read := private.Group("xy/account")
	write := private.Group("xy/account")
	read.GET("list", r.api.Account.List)
	write.POST("create", r.api.Account.Create)
	write.PUT("update", r.api.Account.Update)
	write.DELETE("delete", r.api.Account.Delete)
	read.PUT("credential", r.api.Account.SaveCredential)
	write.POST("check-login", r.api.Account.CheckLogin)
}
