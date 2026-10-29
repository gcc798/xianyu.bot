package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type account struct{}

func (r *account) Init(private *gin.RouterGroup) {
	read := private.Group("xy/account")
	write := private.Group("xy/account").Use(middleware.OperationRecord())
	read.GET("list", accountApi.List)
	write.POST("create", accountApi.Create)
	write.PUT("update", accountApi.Update)
	write.DELETE("delete", accountApi.Delete)
	// Credential bodies must not enter the operation-record log.
	read.PUT("credential", accountApi.SaveCredential)
	write.POST("check-login", accountApi.CheckLogin)
}
