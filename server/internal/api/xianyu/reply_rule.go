package api

import (
	"github.com/gcc798/go.xianyu.assistant/server/internal/model/common/response"
	xyRequest "github.com/gcc798/go.xianyu.assistant/server/internal/model/xianyu/request"
	"github.com/gcc798/go.xianyu.assistant/server/internal/security"
	"github.com/gin-gonic/gin"
)

type replyRule struct{ *dependencies }

// ListReplyRules 分页获取回复规则
// @Tags XyReplyRule
// @Summary 分页获取回复规则
// @Security ApiKeyAuth
// @Produce application/json
// @Param data query xyRequest.ReplyRuleSearch true "分页与筛选条件"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /xy/reply-rule/list [get]
func (a *replyRule) List(c *gin.Context) {
	var input xyRequest.ReplyRuleSearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := a.services.ReplyRule.List(c.Request.Context(), security.GetUserID(c), input)
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
func (a *replyRule) Save(c *gin.Context) {
	var input xyRequest.SaveReplyRule
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := a.services.ReplyRule.Save(c.Request.Context(), security.GetUserID(c), input); err != nil {
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
func (a *replyRule) Delete(c *gin.Context) {
	var input xyRequest.ReplyRuleID
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := a.services.ReplyRule.Delete(c.Request.Context(), security.GetUserID(c), input.ID); err != nil {
		failAccount("删除回复规则失败", err, c)
		return
	}
	response.OkWithMessage("删除成功", c)
}
