package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

type AccountRouter struct{ handlers *api.API }

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
