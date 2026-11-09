package api

import (
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gcc798/xianyu.bot/internal/service"
	"github.com/gin-gonic/gin"
)

// ReplyRuleAPI 处理回复规则 HTTP 请求。
type ReplyRuleAPI struct {
	service *service.Service // service 提供业务能力。
}

// NewReplyRuleAPI 创建 ReplyRuleAPI。
func NewReplyRuleAPI(services *service.Service) *ReplyRuleAPI {
	return &ReplyRuleAPI{service: services}
}

// ListReplyRules 分页获取回复规则
// @Tags XyReplyRule
// @Summary 分页获取回复规则
// @Security ApiKeyAuth
// @Produce application/json
// @Param data query xyRequest.ReplyRuleSearch true "分页与筛选条件"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /xy/reply-rule/list [get]
func (a *ReplyRuleAPI) List(c *gin.Context) {
	var input xyRequest.ReplyRuleSearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := a.service.ReplyRule.List(c.Request.Context(), security.GetUserID(c), input)
	if err != nil {
		failAccount("获取回复规则失败", err, c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: input.Page, PageSize: input.PageSize}, "获取成功", c)
}

// SaveReplyRule 创建或更新回复规则
// @Tags XyReplyRule
// @Summary 创建或更新回复规则
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body xyRequest.SaveReplyRule true "回复规则"
// @Success 200 {object} response.Response{msg=string} "保存成功"
// @Router /xy/reply-rule/save [post]
func (a *ReplyRuleAPI) Save(c *gin.Context) {
	var input xyRequest.SaveReplyRule
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := a.service.ReplyRule.Save(c.Request.Context(), security.GetUserID(c), input); err != nil {
		failAccount("保存回复规则失败", err, c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// DeleteReplyRule 删除回复规则
// @Tags XyReplyRule
// @Summary 删除回复规则
// @Security ApiKeyAuth
// @Produce application/json
// @Param ID query int true "规则 ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /xy/reply-rule/delete [delete]
func (a *ReplyRuleAPI) Delete(c *gin.Context) {
	var input xyRequest.ReplyRuleID
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := a.service.ReplyRule.Delete(c.Request.Context(), security.GetUserID(c), input.ID); err != nil {
		failAccount("删除回复规则失败", err, c)
		return
	}
	response.OkWithMessage("删除成功", c)
}
