package api

import (
	"github.com/gcc798/go.xianyu.assistant/server/internal/infra"
	"github.com/gcc798/go.xianyu.assistant/server/internal/model/common/response"
	xyRequest "github.com/gcc798/go.xianyu.assistant/server/internal/model/xianyu/request"
	"github.com/gcc798/go.xianyu.assistant/server/internal/security"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type account struct{ *dependencies }

// Create 创建闲鱼账号档案
// @Tags XyAccount
// @Summary 创建闲鱼账号档案
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body xyRequest.CreateAccount true "账号资料"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /xy/account/create [post]
func (a *account) Create(c *gin.Context) {
	var input xyRequest.CreateAccount
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	item, err := a.services.Account.Create(c.Request.Context(), security.GetUserID(c), input)
	if err != nil {
		failAccount("创建闲鱼账号失败", err, c)
		return
	}
	response.OkWithDetailed(gin.H{"account": item}, "创建成功", c)
}

// Update 更新闲鱼账号档案
// @Tags XyAccount
// @Summary 更新闲鱼账号档案
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body xyRequest.UpdateAccount true "账号资料"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /xy/account/update [put]
func (a *account) Update(c *gin.Context) {
	var input xyRequest.UpdateAccount
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := a.services.Account.Update(c.Request.Context(), security.GetUserID(c), input); err != nil {
		failAccount("更新闲鱼账号失败", err, c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// Delete 删除闲鱼账号档案
// @Tags XyAccount
// @Summary 删除闲鱼账号档案
// @Security ApiKeyAuth
// @Produce application/json
// @Param ID query int true "账号 ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /xy/account/delete [delete]
func (a *account) Delete(c *gin.Context) {
	var input xyRequest.AccountID
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := a.services.Account.Delete(c.Request.Context(), security.GetUserID(c), input.ID); err != nil {
		failAccount("删除闲鱼账号失败", err, c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// List 分页获取当前用户的闲鱼账号
// @Tags XyAccount
// @Summary 分页获取当前用户的闲鱼账号
// @Security ApiKeyAuth
// @Produce application/json
// @Param data query xyRequest.AccountSearch true "分页与筛选条件"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /xy/account/list [get]
func (a *account) List(c *gin.Context) {
	var input xyRequest.AccountSearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := a.services.Account.List(c.Request.Context(), security.GetUserID(c), input)
	if err != nil {
		failAccount("获取闲鱼账号失败", err, c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: input.Page, PageSize: input.PageSize}, "获取成功", c)
}

// SaveCredential 加密保存闲鱼账号凭证
// @Tags XyAccount
// @Summary 加密保存闲鱼账号凭证
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body xyRequest.SaveCredential true "账号凭证"
// @Success 200 {object} response.Response{msg=string} "保存成功"
// @Router /xy/account/credential [put]
func (a *account) SaveCredential(c *gin.Context) {
	var input xyRequest.SaveCredential
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := a.services.Account.SaveCredential(c.Request.Context(), security.GetUserID(c), input); err != nil {
		failAccount("保存闲鱼账号凭证失败", err, c)
		return
	}
	response.OkWithMessage("凭证已加密保存", c)
}

// CheckLogin 检查登录态并加密保存响应中的 Cookie 更新
// @Tags XyAccount
// @Summary 检查闲鱼账号登录态
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body xyRequest.AccountID true "账号 ID"
// @Success 200 {object} response.Response{data=object,msg=string} "检查完成"
// @Router /xy/account/check-login [post]
func (a *account) CheckLogin(c *gin.Context) {
	var input xyRequest.AccountID
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	status, err := a.services.Account.CheckLogin(c.Request.Context(), security.GetUserID(c), input.ID)
	if err != nil {
		failAccount("检查闲鱼登录状态失败", err, c)
		return
	}
	response.OkWithDetailed(gin.H{"login": status}, "检查完成", c)
}

func failAccount(message string, err error, c *gin.Context) {
	infra.Log.Error(message, zap.Error(err))
	response.FailWithMessage(message, c)
}
