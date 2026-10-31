package api

import (
	xyRequest "github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
)

type deliveryRule struct{}

// ListDeliveryRules 分页获取发货规则
// @Tags XyDeliveryRule
// @Summary 分页获取发货规则
// @Security ApiKeyAuth
// @Produce application/json
// @Param data query xyRequest.DeliveryRuleSearch true "分页与筛选条件"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /xy/delivery-rule/list [get]
func (a *deliveryRule) List(c *gin.Context) {
	var input xyRequest.DeliveryRuleSearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := deliveryRuleService.List(c.Request.Context(), utils.GetUserID(c), input)
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
func (a *deliveryRule) Save(c *gin.Context) {
	var input xyRequest.SaveDeliveryRule
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := deliveryRuleService.Save(c.Request.Context(), utils.GetUserID(c), input); err != nil {
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
func (a *deliveryRule) Delete(c *gin.Context) {
	var input xyRequest.DeliveryRuleID
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := deliveryRuleService.Delete(c.Request.Context(), utils.GetUserID(c), input.ID); err != nil {
		failAccount("删除发货规则失败", err, c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// ImportResources 导入加密卡密库存
func (a *deliveryRule) ImportResources(c *gin.Context) {
	var input xyRequest.ImportDeliveryResources
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	count, err := deliveryService.ImportResources(c.Request.Context(), utils.GetUserID(c), input)
	if err != nil {
		failAccount("导入卡密失败", err, c)
		return
	}
	response.OkWithData(gin.H{"count": count}, c)
}

// Inventory 获取卡密库存统计
func (a *deliveryRule) Inventory(c *gin.Context) {
	var input xyRequest.DeliveryRuleID
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	data, err := deliveryService.Inventory(c.Request.Context(), utils.GetUserID(c), input.ID)
	if err != nil {
		failAccount("获取卡密库存失败", err, c)
		return
	}
	response.OkWithData(data, c)
}

// ListTasks 分页获取自动发货任务
func (a *deliveryRule) ListTasks(c *gin.Context) {
	var input xyRequest.DeliveryTaskSearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := deliveryService.ListTasks(c.Request.Context(), utils.GetUserID(c), input)
	if err != nil {
		failAccount("获取发货任务失败", err, c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: input.Page, PageSize: input.PageSize}, "获取成功", c)
}

// RetryTask 重试明确失败的自动发货任务
func (a *deliveryRule) RetryTask(c *gin.Context) {
	var input xyRequest.DeliveryTaskID
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := deliveryService.Retry(c.Request.Context(), utils.GetUserID(c), input.ID); err != nil {
		failAccount("重试发货失败", err, c)
		return
	}
	response.OkWithMessage("发货任务已执行", c)
}
