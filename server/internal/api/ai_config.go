package api

import (
	"strconv"

	"github.com/gcc798/xianyu.bot/internal/agent"
	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AIConfigAPI 处理Agent 配置 HTTP 请求。
type AIConfigAPI struct {
	config *agent.AIConfig // config 提供 Agent 配置能力。
}

// aiModelConfig 是仅管理员配置接口使用的完整模型配置。
type aiModelConfig struct {
	model.AIModel
	APIKey string `json:"apiKey"` // APIKey 是管理员编辑模型时展示的完整凭据。
}

// NewAIConfigAPI 创建 AIConfigAPI。
func NewAIConfigAPI(config *agent.AIConfig) *AIConfigAPI {
	return &AIConfigAPI{config: config}
}

// ListModels 处理模型配置列表查询请求。
func (a *AIConfigAPI) ListModels(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	items, err := a.config.ListModels(c.Request.Context())
	if err != nil {
		failAIConfig("获取模型配置失败", err, c)
		return
	}
	configs := make([]aiModelConfig, len(items))
	for i := range items {
		configs[i] = aiModelConfig{AIModel: items[i], APIKey: items[i].APIKey}
	}
	response.OkWithData(gin.H{"list": configs}, c)
}

// CreateModel 处理模型配置创建请求。
func (a *AIConfigAPI) CreateModel(c *gin.Context) {
	var input xyRequest.SaveAIModel
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	item, err := a.config.CreateModel(c.Request.Context(), input)
	if err != nil {
		failAIConfig("创建模型配置失败", err, c)
		return
	}
	response.OkWithData(gin.H{"model": item}, c)
}

// UpdateModel 处理模型配置更新请求。
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
	if err := a.config.UpdateModel(c.Request.Context(), id, input); err != nil {
		failAIConfig("更新模型配置失败", err, c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// DeleteModel 处理模型配置删除请求。
func (a *AIConfigAPI) DeleteModel(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := a.config.DeleteModel(c.Request.Context(), id); err != nil {
		failAIConfig("删除模型配置失败", err, c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// ListMCPServers 处理 MCP 服务列表查询请求。
func (a *AIConfigAPI) ListMCPServers(c *gin.Context) {
	items, err := a.config.ListMCPServers(c.Request.Context())
	if err != nil {
		failAIConfig("获取 MCP 服务配置失败", err, c)
		return
	}
	response.OkWithData(gin.H{"list": items}, c)
}

// CreateMCPServer 处理 MCP 服务创建请求。
func (a *AIConfigAPI) CreateMCPServer(c *gin.Context) {
	var input xyRequest.SaveMCPServer
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	item, err := a.config.CreateMCPServer(c.Request.Context(), input)
	if err != nil {
		failAIConfig("创建 MCP 服务配置失败", err, c)
		return
	}
	response.OkWithData(gin.H{"server": item}, c)
}

// UpdateMCPServer 处理 MCP 服务更新请求。
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
	if err := a.config.UpdateMCPServer(c.Request.Context(), id, input); err != nil {
		failAIConfig("更新 MCP 服务配置失败", err, c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// DeleteMCPServer 处理 MCP 服务删除请求。
func (a *AIConfigAPI) DeleteMCPServer(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := a.config.DeleteMCPServer(c.Request.Context(), id); err != nil {
		failAIConfig("删除 MCP 服务配置失败", err, c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetAgentConfig 处理 Agent 配置查询请求。
func (a *AIConfigAPI) GetAgentConfig(c *gin.Context) {
	item, err := a.config.GetAgentConfig(c.Request.Context())
	if err != nil {
		failAIConfig("获取 Agent 配置失败", err, c)
		return
	}
	response.OkWithData(gin.H{"config": item}, c)
}

// UpdateAgentConfig 处理 Agent 配置更新请求。
func (a *AIConfigAPI) UpdateAgentConfig(c *gin.Context) {
	var input model.AgentConfig
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := a.config.UpdateAgentConfig(c.Request.Context(), input); err != nil {
		failAIConfig("更新 Agent 配置失败", err, c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// pathID 解析并校验路径中的资源标识。
func pathID(c *gin.Context) (uint, bool) {
	value, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || value == 0 {
		response.FailWithMessage("ID 无效", c)
		return 0, false
	}
	return uint(value), true
}

// failAIConfig 统一返回 AI 配置操作错误。
func failAIConfig(message string, err error, c *gin.Context) {
	zap.L().Error(message, zap.Error(err))
	response.FailWithMessage(message+"："+err.Error(), c)
}
