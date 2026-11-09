package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gcc798/xianyu.bot/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// DictionaryDetailRouter 注册字典项 HTTP 路由。
type DictionaryDetailRouter struct {
	handlers *api.API // handlers 提供 HTTP 处理器。
}

// NewDictionaryDetailRouter 创建 DictionaryDetailRouter。
func NewDictionaryDetailRouter(handlers *api.API) *DictionaryDetailRouter {
	return &DictionaryDetailRouter{handlers: handlers}
}

// Init 注册当前模块路由。
func (s *DictionaryDetailRouter) Init(router *gin.RouterGroup) {
	detail := router.Group("sysDictionaryDetail").Use(middleware.AdminOnly())
	detail.POST("createSysDictionaryDetail", s.handlers.DictionaryDetail.Create)
	detail.DELETE("deleteSysDictionaryDetail", s.handlers.DictionaryDetail.Delete)
	detail.PUT("updateSysDictionaryDetail", s.handlers.DictionaryDetail.Update)
	detail.GET("getSysDictionaryDetailList", s.handlers.DictionaryDetail.List)
}
