package router

import "github.com/gin-gonic/gin"

type account struct{}

func (r *account) Init(private *gin.RouterGroup) {
	read := private.Group("xy/account")
	write := private.Group("xy/account")
	read.GET("list", accountApi.List)
	write.POST("create", accountApi.Create)
	write.PUT("update", accountApi.Update)
	write.DELETE("delete", accountApi.Delete)
	read.PUT("credential", accountApi.SaveCredential)
	write.POST("check-login", accountApi.CheckLogin)
}
