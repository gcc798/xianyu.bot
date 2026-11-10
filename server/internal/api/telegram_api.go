package api

import (
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gcc798/xianyu.bot/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TelegramAPIConfigAPI 管理 Telegram 应用凭据。
type TelegramAPIConfigAPI struct {
	config *service.TelegramAPIConfig
}

// telegramAPIInternalCredential 只用于 Python 内网服务读取当前凭据。
type telegramAPIInternalCredential struct {
	APIID   int64  `json:"apiId"`
	APIHash string `json:"apiHash"`
}

// NewTelegramAPIConfigAPI 创建 TelegramAPIConfigAPI。
func NewTelegramAPIConfigAPI(config *service.TelegramAPIConfig) *TelegramAPIConfigAPI {
	return &TelegramAPIConfigAPI{config: config}
}

// List 返回分页配置列表，不返回 API Hash。
func (a *TelegramAPIConfigAPI) List(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	var input xyRequest.TelegramAPIConfigSearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage("分页或搜索参数无效", c)
		return
	}
	configs, total, err := a.config.List(c.Request.Context(), input)
	if err != nil {
		failTelegramAPI("获取 Telegram API 配置失败", err, c)
		return
	}
	response.OkWithData(gin.H{"page": input.Page, "pageSize": input.PageSize, "total": total, "list": configs}, c)
}

// Create 新增 Telegram API 配置。
func (a *TelegramAPIConfigAPI) Create(c *gin.Context) {
	var input xyRequest.SaveTelegramAPIConfig
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	config, err := a.config.Create(c.Request.Context(), input)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	config.APIHash = ""
	response.OkWithData(gin.H{"config": config}, c)
}

// Update 修改 Telegram API 配置。
func (a *TelegramAPIConfigAPI) Update(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var input xyRequest.SaveTelegramAPIConfig
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := a.config.Update(c.Request.Context(), id, input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功，Telegram 服务将在下次任务开始时采用可用配置", c)
}

// Delete 删除 Telegram API 配置。
func (a *TelegramAPIConfigAPI) Delete(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := a.config.Delete(c.Request.Context(), id); err != nil {
		failTelegramAPI("删除 Telegram API 配置失败", err, c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// ActiveCredential 返回内部 Telegram 服务当前可用的最高 API ID。
func (a *TelegramAPIConfigAPI) ActiveCredential(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	config, err := a.config.Active(c.Request.Context())
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(telegramAPIInternalCredential{APIID: config.APIID, APIHash: config.APIHash}, c)
}

// ReportFloodWait 记录内部 Telegram 服务上报的限流时长。
func (a *TelegramAPIConfigAPI) ReportFloodWait(c *gin.Context) {
	var input xyRequest.ReportTelegramFloodWait
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := a.config.RecordFloodWait(c.Request.Context(), input.APIID, input.Seconds); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("限流状态已更新", c)
}

func failTelegramAPI(message string, err error, c *gin.Context) {
	zap.L().Error(message, zap.Error(err))
	response.FailWithMessage(message, c)
}
