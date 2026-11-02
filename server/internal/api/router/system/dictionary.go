package system

import (
	"github.com/gcc798/xianyu.bot/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

type DictionaryRouter struct{ *dependencies }

func (s *DictionaryRouter) Init(router *gin.RouterGroup) {
	dictionary := router.Group("sysDictionary").Use(middleware.AdminOnly())
	dictionary.POST("createSysDictionary", s.handlers.Dictionary.Create)
	dictionary.DELETE("deleteSysDictionary", s.handlers.Dictionary.Delete)
	dictionary.PUT("updateSysDictionary", s.handlers.Dictionary.Update)
	dictionary.GET("getSysDictionaryList", s.handlers.Dictionary.List)
}
