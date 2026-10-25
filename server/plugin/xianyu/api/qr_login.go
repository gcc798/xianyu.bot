package api

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	xianyuRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/xianyu/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
)

type qrLogin struct{}

// Start 创建账号二维码登录会话
// @Tags XianyuQRLogin
// @Summary 创建账号二维码登录会话
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body xianyuRequest.AccountID true "账号 ID"
// @Success 200 {object} response.Response{data=object,msg=string} "创建成功"
// @Router /xianyu/qr-login/start [post]
func (a *qrLogin) Start(c *gin.Context) {
	var input xianyuRequest.AccountID
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	login, err := qrLoginService.Start(c.Request.Context(), utils.GetUserID(c), input.ID)
	if err != nil {
		failAccount("创建闲鱼二维码失败", err, c)
		return
	}
	response.OkWithDetailed(gin.H{"login": login}, "二维码已生成", c)
}

// Status 获取二维码登录状态
// @Tags XianyuQRLogin
// @Summary 获取二维码登录状态
// @Security ApiKeyAuth
// @Produce application/json
// @Param sessionId query string true "会话 ID"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /xianyu/qr-login/status [get]
func (a *qrLogin) Status(c *gin.Context) {
	var input xianyuRequest.QRSessionID
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	login, err := qrLoginService.Status(utils.GetUserID(c), input.SessionID)
	if err != nil {
		failAccount("获取二维码登录状态失败", err, c)
		return
	}
	response.OkWithDetailed(gin.H{"login": login}, "获取成功", c)
}

// Cancel 取消二维码登录会话
// @Tags XianyuQRLogin
// @Summary 取消二维码登录会话
// @Security ApiKeyAuth
// @Produce application/json
// @Param sessionId query string true "会话 ID"
// @Success 200 {object} response.Response{msg=string} "取消成功"
// @Router /xianyu/qr-login/cancel [delete]
func (a *qrLogin) Cancel(c *gin.Context) {
	var input xianyuRequest.QRSessionID
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := qrLoginService.Cancel(utils.GetUserID(c), input.SessionID); err != nil {
		failAccount("取消二维码登录失败", err, c)
		return
	}
	response.OkWithMessage("登录已取消", c)
}
