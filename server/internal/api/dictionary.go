package api

import (
	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gcc798/xianyu.bot/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DictionaryAPI struct{ service *service.Service }

func NewDictionaryAPI(services *service.Service) *DictionaryAPI {
	return &DictionaryAPI{service: services}
}

func (s *DictionaryAPI) Create(c *gin.Context) {
	var dictionary model.Dictionary
	if err := c.ShouldBindJSON(&dictionary); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := s.service.Dictionary.Create(dictionary); err != nil {
		zap.L().Error("创建字典失败", zap.Error(err))
		response.FailWithMessage("创建失败", c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

func (s *DictionaryAPI) Delete(c *gin.Context) {
	var dictionary model.Dictionary
	if err := c.ShouldBindJSON(&dictionary); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := s.service.Dictionary.Delete(dictionary); err != nil {
		zap.L().Error("删除字典失败", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (s *DictionaryAPI) Update(c *gin.Context) {
	var dictionary model.Dictionary
	if err := c.ShouldBindJSON(&dictionary); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := s.service.Dictionary.Update(&dictionary); err != nil {
		zap.L().Error("更新字典失败", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (s *DictionaryAPI) List(c *gin.Context) {
	var search request.DictionarySearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := s.service.Dictionary.List(c.Request.Context(), search)
	if err != nil {
		zap.L().Error("获取字典失败", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
