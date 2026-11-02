package api

import (
	"github.com/gcc798/go.xianyu.assistant/server/internal/model/common/response"
	xyRequest "github.com/gcc798/go.xianyu.assistant/server/internal/model/xianyu/request"
	"github.com/gcc798/go.xianyu.assistant/server/internal/security"
	"github.com/gin-gonic/gin"
)

type qrLogin struct{ *dependencies }

// Start 创建账号二维码登录会话
// @Tags XyQRLogin
// @Summary 创建账号二维码登录会话
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body xyRequest.AccountID true "账号 ID"
// @Success 200 {object} response.Response{data=object,msg=string} "创建成功"
// @Router /xy/qr-login/start [post]
func (a *qrLogin) Start(c *gin.Context) {
	var input xyRequest.AccountID
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	login, err := a.services.QRLogin.Start(c.Request.Context(), security.GetUserID(c), input.ID)
	if err != nil {
		failAccount("创建闲鱼二维码失败", err, c)
		return
	}
	response.OkWithDetailed(gin.H{"login": login}, "二维码已生成", c)
}

// Status 获取二维码登录状态
// @Tags XyQRLogin
// @Summary 获取二维码登录状态
// @Security ApiKeyAuth
// @Produce application/json
// @Param sessionId query string true "会话 ID"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /xy/qr-login/status [get]
func (a *qrLogin) Status(c *gin.Context) {
	var input xyRequest.QRSessionID
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	login, err := a.services.QRLogin.Status(security.GetUserID(c), input.SessionID)
	if err != nil {
		failAccount("获取二维码登录状态失败", err, c)
		return
	}
	response.OkWithDetailed(gin.H{"login": login}, "获取成功", c)
}

// Cancel 取消二维码登录会话
// @Tags XyQRLogin
// @Summary 取消二维码登录会话
// @Security ApiKeyAuth
// @Produce application/json
// @Param sessionId query string true "会话 ID"
// @Success 200 {object} response.Response{msg=string} "取消成功"
// @Router /xy/qr-login/cancel [delete]
func (a *qrLogin) Cancel(c *gin.Context) {
	var input xyRequest.QRSessionID
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := a.services.QRLogin.Cancel(security.GetUserID(c), input.SessionID); err != nil {
		failAccount("取消二维码登录失败", err, c)
		return
	}
	response.OkWithMessage("登录已取消", c)
}
