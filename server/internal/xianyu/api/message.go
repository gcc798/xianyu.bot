package api

import (
	xyRequest "github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
)

type message struct{}

// List 分页获取当前用户的闲鱼消息
// @Tags XyMessage
// @Summary 分页获取当前用户的闲鱼消息
// @Security ApiKeyAuth
// @Produce application/json
// @Param data query xyRequest.MessageSearch true "分页与筛选条件"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /xy/message/list [get]
func (a *message) List(c *gin.Context) {
	var input xyRequest.MessageSearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := messageService.List(c.Request.Context(), utils.GetUserID(c), input)
	if err != nil {
		failAccount("获取闲鱼消息失败", err, c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: input.Page, PageSize: input.PageSize}, "获取成功", c)
}

// Reply 在指定会话中发送人工文本回复
// @Tags XyMessage
// @Summary 发送人工文本回复
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body xyRequest.SendMessage true "回复内容"
// @Success 200 {object} response.Response{msg=string} "发送成功"
// @Router /xy/message/reply [post]
func (a *message) Reply(c *gin.Context) {
	var input xyRequest.SendMessage
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := messageService.Reply(c.Request.Context(), utils.GetUserID(c), input); err != nil {
		failAccount("发送闲鱼消息失败", err, c)
		return
	}
	response.OkWithMessage("发送成功，当前会话已转为人工接管", c)
}

// UpdateTakeover 更新会话人工接管状态
// @Tags XyMessage
// @Summary 更新会话人工接管状态
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body xyRequest.UpdateTakeover true "接管状态"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /xy/message/takeover [put]
func (a *message) UpdateTakeover(c *gin.Context) {
	var input xyRequest.UpdateTakeover
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := messageService.UpdateTakeover(c.Request.Context(), utils.GetUserID(c), input); err != nil {
		failAccount("更新人工接管状态失败", err, c)
		return
	}
	response.OkWithMessage("更新成功", c)
}
