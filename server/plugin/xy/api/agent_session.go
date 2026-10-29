package api

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	xyRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
)

type agentSession struct{}

func (a *agentSession) Create(c *gin.Context) {
	var input xyRequest.CreateAgentSession
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	session, err := agentSessionService.Create(c.Request.Context(), utils.GetUserID(c), input)
	if err != nil {
		failAccount("创建 Agent 会话失败", err, c)
		return
	}
	response.OkWithData(gin.H{"session": session}, c)
}

func (a *agentSession) List(c *gin.Context) {
	var input xyRequest.AgentSessionSearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	sessions, total, err := agentSessionService.List(c.Request.Context(), utils.GetUserID(c), input)
	if err != nil {
		failAccount("获取 Agent 会话失败", err, c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: sessions, Total: total, Page: input.Page, PageSize: input.PageSize}, "获取成功", c)
}

func (a *agentSession) Get(c *gin.Context) {
	sessionID, ok := pathID(c)
	if !ok {
		return
	}
	session, err := agentSessionService.Get(c.Request.Context(), utils.GetUserID(c), sessionID)
	if err != nil {
		failAccount("获取 Agent 会话详情失败", err, c)
		return
	}
	response.OkWithData(gin.H{"session": session}, c)
}

func (a *agentSession) Delete(c *gin.Context) {
	sessionID, ok := pathID(c)
	if !ok {
		return
	}
	if err := agentSessionService.Delete(c.Request.Context(), utils.GetUserID(c), sessionID); err != nil {
		failAccount("删除 Agent 会话失败", err, c)
		return
	}
	response.OkWithMessage("删除成功", c)
}
