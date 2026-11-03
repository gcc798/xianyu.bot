package api

import (
	"strconv"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gcc798/xianyu.bot/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AIConfigAPI struct{ service *service.Service }

func (a *AIConfigAPI) ListModels(c *gin.Context) {
	items, err := a.service.AIConfig.ListModels(c.Request.Context())
	if err != nil {
		failAIConfig("获取模型配置失败", err, c)
		return
	}
	response.OkWithData(gin.H{"list": items}, c)
}

func (a *AIConfigAPI) CreateModel(c *gin.Context) {
	var input xyRequest.SaveAIModel
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	item, err := a.service.AIConfig.CreateModel(c.Request.Context(), input)
	if err != nil {
		failAIConfig("创建模型配置失败", err, c)
		return
	}
	response.OkWithData(gin.H{"model": item}, c)
}

func (a *AIConfigAPI) UpdateModel(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var input xyRequest.SaveAIModel
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := a.service.AIConfig.UpdateModel(c.Request.Context(), id, input); err != nil {
		failAIConfig("更新模型配置失败", err, c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

func (a *AIConfigAPI) DeleteModel(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := a.service.AIConfig.DeleteModel(c.Request.Context(), id); err != nil {
		failAIConfig("删除模型配置失败", err, c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (a *AIConfigAPI) SetDefaultModel(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := a.service.AIConfig.SetDefaultModel(c.Request.Context(), id); err != nil {
		failAIConfig("设置默认模型失败", err, c)
		return
	}
	response.OkWithMessage("默认模型已更新", c)
}

func (a *AIConfigAPI) ListMCPServers(c *gin.Context) {
	items, err := a.service.AIConfig.ListMCPServers(c.Request.Context())
	if err != nil {
		failAIConfig("获取 MCP 服务配置失败", err, c)
		return
	}
	response.OkWithData(gin.H{"list": items}, c)
}

func (a *AIConfigAPI) CreateMCPServer(c *gin.Context) {
	var input xyRequest.SaveMCPServer
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	item, err := a.service.AIConfig.CreateMCPServer(c.Request.Context(), input)
	if err != nil {
		failAIConfig("创建 MCP 服务配置失败", err, c)
		return
	}
	response.OkWithData(gin.H{"server": item}, c)
}

func (a *AIConfigAPI) UpdateMCPServer(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var input xyRequest.SaveMCPServer
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := a.service.AIConfig.UpdateMCPServer(c.Request.Context(), id, input); err != nil {
		failAIConfig("更新 MCP 服务配置失败", err, c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

func (a *AIConfigAPI) DeleteMCPServer(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := a.service.AIConfig.DeleteMCPServer(c.Request.Context(), id); err != nil {
		failAIConfig("删除 MCP 服务配置失败", err, c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (a *AIConfigAPI) GetAgentConfig(c *gin.Context) {
	item, err := a.service.AIConfig.GetAgentConfig(c.Request.Context())
	if err != nil {
		failAIConfig("获取 Agent 配置失败", err, c)
		return
	}
	response.OkWithData(gin.H{"config": item}, c)
}

func (a *AIConfigAPI) UpdateAgentConfig(c *gin.Context) {
	var input model.AgentConfig
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := a.service.AIConfig.UpdateAgentConfig(c.Request.Context(), input); err != nil {
		failAIConfig("更新 Agent 配置失败", err, c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

func pathID(c *gin.Context) (uint, bool) {
	value, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || value == 0 {
		response.FailWithMessage("ID 无效", c)
		return 0, false
	}
	return uint(value), true
}

func failAIConfig(message string, err error, c *gin.Context) {
	zap.L().Error(message, zap.Error(err))
	response.FailWithMessage(message+"："+err.Error(), c)
}
