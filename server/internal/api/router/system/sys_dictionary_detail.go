package system

import (
	"github.com/gcc798/go.xianyu.assistant/server/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

type DictionaryDetailRouter struct{ *dependencies }

func (s *DictionaryDetailRouter) InitSysDictionaryDetailRouter(router *gin.RouterGroup) {
	detail := router.Group("sysDictionaryDetail").Use(middleware.AdminOnly())
	detail.POST("createSysDictionaryDetail", s.handlers.DictionaryDetail.CreateSysDictionaryDetail)
	detail.DELETE("deleteSysDictionaryDetail", s.handlers.DictionaryDetail.DeleteSysDictionaryDetail)
	detail.PUT("updateSysDictionaryDetail", s.handlers.DictionaryDetail.UpdateSysDictionaryDetail)
	detail.GET("getSysDictionaryDetailList", s.handlers.DictionaryDetail.GetSysDictionaryDetailList)
}
