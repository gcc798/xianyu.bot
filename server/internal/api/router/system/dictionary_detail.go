package system

import (
	"github.com/gcc798/xianyu.bot/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

type DictionaryDetailRouter struct{ *dependencies }

func (s *DictionaryDetailRouter) Init(router *gin.RouterGroup) {
	detail := router.Group("sysDictionaryDetail").Use(middleware.AdminOnly())
	detail.POST("createSysDictionaryDetail", s.handlers.DictionaryDetail.Create)
	detail.DELETE("deleteSysDictionaryDetail", s.handlers.DictionaryDetail.Delete)
	detail.PUT("updateSysDictionaryDetail", s.handlers.DictionaryDetail.Update)
	detail.GET("getSysDictionaryDetailList", s.handlers.DictionaryDetail.List)
}
