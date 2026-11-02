package system

import (
	"github.com/gcc798/xianyu.bot/internal/model/common/response"
	"github.com/gcc798/xianyu.bot/internal/model/system"
	"github.com/gcc798/xianyu.bot/internal/model/system/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DictionaryDetailApi struct{ *dependencies }

func (s *DictionaryDetailApi) Create(c *gin.Context) {
	var detail system.DictionaryDetail
	if err := c.ShouldBindJSON(&detail); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := s.services.DictionaryDetail.Create(detail); err != nil {
		zap.L().Error("创建字典项失败", zap.Error(err))
		response.FailWithMessage("创建失败", c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

func (s *DictionaryDetailApi) Delete(c *gin.Context) {
	var detail system.DictionaryDetail
	if err := c.ShouldBindJSON(&detail); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := s.services.DictionaryDetail.Delete(detail); err != nil {
		zap.L().Error("删除字典项失败", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (s *DictionaryDetailApi) Update(c *gin.Context) {
	var detail system.DictionaryDetail
	if err := c.ShouldBindJSON(&detail); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := s.services.DictionaryDetail.Update(&detail); err != nil {
		zap.L().Error("更新字典项失败", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (s *DictionaryDetailApi) List(c *gin.Context) {
	var search request.DictionaryDetailSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := s.services.DictionaryDetail.List(search)
	if err != nil {
		zap.L().Error("获取字典项失败", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List: list, Total: total, Page: search.Page, PageSize: search.PageSize,
	}, "获取成功", c)
}
