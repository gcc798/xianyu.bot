package api

import (
	xyRequest "github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
)

type webSocket struct{}

// Start 启动闲鱼账号消息连接
// @Tags XyWebSocket
// @Summary 启动闲鱼账号消息连接
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body xyRequest.AccountID true "账号 ID"
// @Success 200 {object} response.Response{msg=string} "启动成功"
// @Router /xy/websocket/start [post]
func (a *webSocket) Start(c *gin.Context) {
	var input xyRequest.AccountID
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := webSocketService.Start(c.Request.Context(), utils.GetUserID(c), input.ID); err != nil {
		failAccount("启动闲鱼消息连接失败", err, c)
		return
	}
	response.OkWithMessage("消息连接正在启动", c)
}

// Stop 停止闲鱼账号消息连接
// @Tags XyWebSocket
// @Summary 停止闲鱼账号消息连接
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body xyRequest.AccountID true "账号 ID"
// @Success 200 {object} response.Response{msg=string} "停止成功"
// @Router /xy/websocket/stop [post]
func (a *webSocket) Stop(c *gin.Context) {
	var input xyRequest.AccountID
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := webSocketService.Stop(c.Request.Context(), utils.GetUserID(c), input.ID); err != nil {
		failAccount("停止闲鱼消息连接失败", err, c)
		return
	}
	response.OkWithMessage("消息连接已停止", c)
}

// Status 获取闲鱼账号消息连接状态
// @Tags XyWebSocket
// @Summary 获取闲鱼账号消息连接状态
// @Security ApiKeyAuth
// @Produce application/json
// @Param ID query int true "账号 ID"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /xy/websocket/status [get]
func (a *webSocket) Status(c *gin.Context) {
	var input xyRequest.AccountID
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	status, err := webSocketService.Status(c.Request.Context(), utils.GetUserID(c), input.ID)
	if err != nil {
		failAccount("获取闲鱼消息连接状态失败", err, c)
		return
	}
	response.OkWithDetailed(gin.H{"connection": status}, "获取成功", c)
}
