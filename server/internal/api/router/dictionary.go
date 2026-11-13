package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gcc798/xianyu.bot/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// DictionaryRouter 注册字典 HTTP 路由。
type DictionaryRouter struct {
	api *api.API // api 提供 HTTP 处理器。
}

// NewDictionaryRouter 创建 DictionaryRouter。
func NewDictionaryRouter(api *api.API) *DictionaryRouter {
	return &DictionaryRouter{api: api}
}

// Init 注册当前模块路由。
func (s *DictionaryRouter) Init(router *gin.RouterGroup) {
	dictionary := router.Group("sysDictionary")
	dictionary.GET("options", s.api.Dictionary.Options)
	dictionary.Use(middleware.AdminOnly())
	dictionary.POST("createSysDictionary", s.api.Dictionary.Create)
	dictionary.DELETE("deleteSysDictionary", s.api.Dictionary.Delete)
	dictionary.PUT("updateSysDictionary", s.api.Dictionary.Update)
	dictionary.GET("getSysDictionaryList", s.api.Dictionary.List)
}
