package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gcc798/xianyu.bot/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// DictionaryRouter 注册字典 HTTP 路由。
type DictionaryRouter struct {
	handlers *api.API // handlers 提供 HTTP 处理器。
}

// NewDictionaryRouter 创建 DictionaryRouter。
func NewDictionaryRouter(handlers *api.API) *DictionaryRouter {
	return &DictionaryRouter{handlers: handlers}
}

// Init 注册当前模块路由。
func (s *DictionaryRouter) Init(router *gin.RouterGroup) {
	dictionary := router.Group("sysDictionary").Use(middleware.AdminOnly())
	dictionary.POST("createSysDictionary", s.handlers.Dictionary.Create)
	dictionary.DELETE("deleteSysDictionary", s.handlers.Dictionary.Delete)
	dictionary.PUT("updateSysDictionary", s.handlers.Dictionary.Update)
	dictionary.GET("getSysDictionaryList", s.handlers.Dictionary.List)
}
