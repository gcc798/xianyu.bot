package system

import (
	"github.com/gcc798/go.xianyu.assistant/server/internal/infra"
	"github.com/gcc798/go.xianyu.assistant/server/internal/model/common/response"
	"github.com/gcc798/go.xianyu.assistant/server/internal/model/system"
	"github.com/gcc798/go.xianyu.assistant/server/internal/model/system/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DictionaryDetailApi struct{ *dependencies }

func (s *DictionaryDetailApi) CreateSysDictionaryDetail(c *gin.Context) {
	var detail system.SysDictionaryDetail
	if err := c.ShouldBindJSON(&detail); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := s.services.DictionaryDetail.CreateSysDictionaryDetail(detail); err != nil {
		infra.Log.Error("创建字典项失败", zap.Error(err))
		response.FailWithMessage("创建失败", c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

func (s *DictionaryDetailApi) DeleteSysDictionaryDetail(c *gin.Context) {
	var detail system.SysDictionaryDetail
	if err := c.ShouldBindJSON(&detail); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := s.services.DictionaryDetail.DeleteSysDictionaryDetail(detail); err != nil {
		infra.Log.Error("删除字典项失败", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (s *DictionaryDetailApi) UpdateSysDictionaryDetail(c *gin.Context) {
	var detail system.SysDictionaryDetail
	if err := c.ShouldBindJSON(&detail); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := s.services.DictionaryDetail.UpdateSysDictionaryDetail(&detail); err != nil {
		infra.Log.Error("更新字典项失败", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (s *DictionaryDetailApi) GetSysDictionaryDetailList(c *gin.Context) {
	var search request.SysDictionaryDetailSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := s.services.DictionaryDetail.GetSysDictionaryDetailInfoList(search)
	if err != nil {
		infra.Log.Error("获取字典项失败", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: search.Page, PageSize: search.PageSize,
	}, "获取成功", c)
}
