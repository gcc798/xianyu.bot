package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DictionaryApi struct{}

func (s *DictionaryApi) CreateSysDictionary(c *gin.Context) {
	var dictionary system.SysDictionary
	if err := c.ShouldBindJSON(&dictionary); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := dictionaryService.CreateSysDictionary(dictionary); err != nil {
		global.GVA_LOG.Error("创建字典失败", zap.Error(err))
		response.FailWithMessage("创建失败", c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

func (s *DictionaryApi) DeleteSysDictionary(c *gin.Context) {
	var dictionary system.SysDictionary
	if err := c.ShouldBindJSON(&dictionary); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := dictionaryService.DeleteSysDictionary(dictionary); err != nil {
		global.GVA_LOG.Error("删除字典失败", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (s *DictionaryApi) UpdateSysDictionary(c *gin.Context) {
	var dictionary system.SysDictionary
	if err := c.ShouldBindJSON(&dictionary); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := dictionaryService.UpdateSysDictionary(&dictionary); err != nil {
		global.GVA_LOG.Error("更新字典失败", zap.Error(err))
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (s *DictionaryApi) GetSysDictionaryList(c *gin.Context) {
	var search request.SysDictionarySearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, err := dictionaryService.GetSysDictionaryInfoList(c.Request.Context(), search)
	if err != nil {
		global.GVA_LOG.Error("获取字典失败", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(list, "获取成功", c)
}
