package system

import (
	"github.com/gcc798/go.xianyu.assistant/server/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

type DictionaryRouter struct{ *dependencies }

func (s *DictionaryRouter) InitSysDictionaryRouter(router *gin.RouterGroup) {
	dictionary := router.Group("sysDictionary").Use(middleware.AdminOnly())
	dictionary.POST("createSysDictionary", s.handlers.Dictionary.CreateSysDictionary)
	dictionary.DELETE("deleteSysDictionary", s.handlers.Dictionary.DeleteSysDictionary)
	dictionary.PUT("updateSysDictionary", s.handlers.Dictionary.UpdateSysDictionary)
	dictionary.GET("getSysDictionaryList", s.handlers.Dictionary.GetSysDictionaryList)
}
