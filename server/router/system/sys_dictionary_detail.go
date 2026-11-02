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
	detail.GET("findSysDictionaryDetail", dictionaryDetailApi.FindSysDictionaryDetail)
	detail.GET("getSysDictionaryDetailList", dictionaryDetailApi.GetSysDictionaryDetailList)
	detail.GET("getDictionaryTreeList", dictionaryDetailApi.GetDictionaryTreeList)
	detail.GET("getDictionaryTreeListByType", dictionaryDetailApi.GetDictionaryTreeListByType)
	detail.GET("getDictionaryDetailsByParent", dictionaryDetailApi.GetDictionaryDetailsByParent)
	detail.GET("getDictionaryPath", dictionaryDetailApi.GetDictionaryPath)
}
