package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type DictionaryRouter struct{}

func (s *DictionaryRouter) InitSysDictionaryRouter(router *gin.RouterGroup) {
	dictionary := router.Group("sysDictionary").Use(middleware.AdminOnly())
	dictionary.POST("createSysDictionary", dictionaryApi.CreateSysDictionary)
	dictionary.DELETE("deleteSysDictionary", dictionaryApi.DeleteSysDictionary)
	dictionary.PUT("updateSysDictionary", dictionaryApi.UpdateSysDictionary)
	dictionary.GET("getSysDictionaryList", dictionaryApi.GetSysDictionaryList)
}
