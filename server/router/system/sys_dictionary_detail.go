package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type DictionaryDetailRouter struct{}

func (s *DictionaryDetailRouter) InitSysDictionaryDetailRouter(router *gin.RouterGroup) {
	detail := router.Group("sysDictionaryDetail").Use(middleware.AdminOnly())
	detail.POST("createSysDictionaryDetail", dictionaryDetailApi.CreateSysDictionaryDetail)
	detail.DELETE("deleteSysDictionaryDetail", dictionaryDetailApi.DeleteSysDictionaryDetail)
	detail.PUT("updateSysDictionaryDetail", dictionaryDetailApi.UpdateSysDictionaryDetail)
	detail.GET("getSysDictionaryDetailList", dictionaryDetailApi.GetSysDictionaryDetailList)
}
