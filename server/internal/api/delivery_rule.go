package api

import (
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gcc798/xianyu.bot/internal/service"
	"github.com/gin-gonic/gin"
)

// DeliveryRuleAPI 处理发货规则 HTTP 请求。
type DeliveryRuleAPI struct {
	service *service.Service // service 提供业务能力。
}

// NewDeliveryRuleAPI 创建 DeliveryRuleAPI。
func NewDeliveryRuleAPI(services *service.Service) *DeliveryRuleAPI {
	return &DeliveryRuleAPI{service: services}
}

// ListDeliveryRules 分页获取发货规则
// @Tags XyDeliveryRule
// @Summary 分页获取发货规则
// @Security ApiKeyAuth
// @Produce application/json
// @Param data query xyRequest.DeliveryRuleSearch true "分页与筛选条件"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /xy/delivery-rule/list [get]
func (a *DeliveryRuleAPI) List(c *gin.Context) {
	var input xyRequest.DeliveryRuleSearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := a.service.DeliveryRule.List(c.Request.Context(), security.GetUserID(c), input)
	if err != nil {
		failAccount("获取发货规则失败", err, c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: input.Page, PageSize: input.PageSize}, "获取成功", c)
}

// SaveDeliveryRule 创建或更新发货规则
// @Tags XyDeliveryRule
// @Summary 创建或更新发货规则
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body xyRequest.SaveDeliveryRule true "发货规则"
// @Success 200 {object} response.Response{msg=string} "保存成功"
// @Router /xy/delivery-rule/save [post]
func (a *DeliveryRuleAPI) Save(c *gin.Context) {
	var input xyRequest.SaveDeliveryRule
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := a.service.DeliveryRule.Save(c.Request.Context(), security.GetUserID(c), input); err != nil {
		failAccount("保存发货规则失败", err, c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// DeleteDeliveryRule 删除发货规则
// @Tags XyDeliveryRule
// @Summary 删除发货规则
// @Security ApiKeyAuth
// @Produce application/json
// @Param ID query int true "规则 ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /xy/delivery-rule/delete [delete]
func (a *DeliveryRuleAPI) Delete(c *gin.Context) {
	var input xyRequest.DeliveryRuleID
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := a.service.DeliveryRule.Delete(c.Request.Context(), security.GetUserID(c), input.ID); err != nil {
		failAccount("删除发货规则失败", err, c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// ImportResources 导入加密卡密库存
func (a *DeliveryRuleAPI) ImportResources(c *gin.Context) {
	var input xyRequest.ImportDeliveryResources
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	count, err := a.service.Delivery.ImportResources(c.Request.Context(), security.GetUserID(c), input)
	if err != nil {
		failAccount("导入卡密失败", err, c)
		return
	}
	response.OkWithData(gin.H{"count": count}, c)
}

// Inventory 获取卡密库存统计
func (a *DeliveryRuleAPI) Inventory(c *gin.Context) {
	var input xyRequest.DeliveryRuleID
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	data, err := a.service.Delivery.Inventory(c.Request.Context(), security.GetUserID(c), input.ID)
	if err != nil {
		failAccount("获取卡密库存失败", err, c)
		return
	}
	response.OkWithData(data, c)
}

// ListTasks 分页获取自动发货任务
func (a *DeliveryRuleAPI) ListTasks(c *gin.Context) {
	var input xyRequest.DeliveryTaskSearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := a.service.Delivery.ListTasks(c.Request.Context(), security.GetUserID(c), input)
	if err != nil {
		failAccount("获取发货任务失败", err, c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: input.Page, PageSize: input.PageSize}, "获取成功", c)
}

// RetryTask 重试明确失败的自动发货任务
func (a *DeliveryRuleAPI) RetryTask(c *gin.Context) {
	var input xyRequest.DeliveryTaskID
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := a.service.Delivery.Retry(c.Request.Context(), security.GetUserID(c), input.ID); err != nil {
		failAccount("重试发货失败", err, c)
		return
	}
	response.OkWithMessage("发货任务已执行", c)
}
