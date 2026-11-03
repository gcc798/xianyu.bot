package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gcc798/xianyu.bot/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

type DictionaryDetailRouter struct{ handlers *api.API }

func NewDictionaryDetailRouter(handlers *api.API) *DictionaryDetailRouter {
	return &DictionaryDetailRouter{handlers: handlers}
}

func (s *DictionaryDetailRouter) Init(router *gin.RouterGroup) {
	detail := router.Group("sysDictionaryDetail").Use(middleware.AdminOnly())
	detail.POST("createSysDictionaryDetail", s.handlers.DictionaryDetail.Create)
	detail.DELETE("deleteSysDictionaryDetail", s.handlers.DictionaryDetail.Delete)
	detail.PUT("updateSysDictionaryDetail", s.handlers.DictionaryDetail.Update)
	detail.GET("getSysDictionaryDetailList", s.handlers.DictionaryDetail.List)
}
