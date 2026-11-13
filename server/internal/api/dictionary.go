package api

import (
	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gcc798/xianyu.bot/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DictionaryAPI 处理字典 HTTP 请求。
type DictionaryAPI struct {
	svc *service.Service // svc 提供业务能力。
}

// NewDictionaryAPI 创建 DictionaryAPI。
func NewDictionaryAPI(svc *service.Service) *DictionaryAPI {
	return &DictionaryAPI{svc: svc}
}

// Create 处理创建请求。
func (s *DictionaryAPI) Create(c *gin.Context) {
	var dictionary model.Dictionary
	if err := c.ShouldBindJSON(&dictionary); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := s.svc.Dictionary.Create(dictionary); err != nil {
		zap.L().Error("创建字典失败", zap.Error(err))
		response.FailWithMessage("创建失败", c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// Delete 处理删除请求。
func (s *DictionaryAPI) Delete(c *gin.Context) {
	var dictionary model.Dictionary
	if err := c.ShouldBindJSON(&dictionary); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := s.svc.Dictionary.Delete(dictionary); err != nil {
		zap.L().Error("删除字典失败", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// Update 处理更新请求。
func (s *DictionaryAPI) Update(c *gin.Context) {
	var dictionary model.Dictionary
	if err := c.ShouldBindJSON(&dictionary); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := s.svc.Dictionary.Update(&dictionary); err != nil {
		zap.L().Error("更新字典失败", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// List 处理列表查询请求。
func (s *DictionaryAPI) List(c *gin.Context) {
	var search request.DictionarySearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := s.svc.Dictionary.List(c.Request.Context(), search)
	if err != nil {
		zap.L().Error("获取字典失败", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}

// Options 返回供业务页面使用的字典编码和名称。
func (s *DictionaryAPI) Options(c *gin.Context) {
	options, err := s.svc.Dictionary.Options(c.Request.Context())
	if err != nil {
		zap.L().Error("获取字典选项失败", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithData(options, c)
}
