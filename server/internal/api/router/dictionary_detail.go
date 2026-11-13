package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gcc798/xianyu.bot/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// DictionaryDetailRouter 注册字典项 HTTP 路由。
type DictionaryDetailRouter struct {
	api *api.API // api 提供 HTTP 处理器。
}

// NewDictionaryDetailRouter 创建 DictionaryDetailRouter。
func NewDictionaryDetailRouter(api *api.API) *DictionaryDetailRouter {
	return &DictionaryDetailRouter{api: api}
}

// Init 注册当前模块路由。
func (s *DictionaryDetailRouter) Init(router *gin.RouterGroup) {
	detail := router.Group("sysDictionaryDetail").Use(middleware.AdminOnly())
	detail.POST("createSysDictionaryDetail", s.api.DictionaryDetail.Create)
	detail.DELETE("deleteSysDictionaryDetail", s.api.DictionaryDetail.Delete)
	detail.PUT("updateSysDictionaryDetail", s.api.DictionaryDetail.Update)
	detail.GET("getSysDictionaryDetailList", s.api.DictionaryDetail.List)
}
