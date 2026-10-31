package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	xyRequest "github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model/request"
	xyResponse "github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

func (a *agentSession) Trajectory(c *gin.Context) {
	sessionID, ok := pathID(c)
	if !ok {
		return
	}
	var input xyRequest.AgentTrajectorySearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	page, err := agentSessionService.Trajectory(c.Request.Context(), utils.GetUserID(c), sessionID, input)
	if err != nil {
		failAccount("获取 Agent 轨迹失败", err, c)
		return
	}
	response.OkWithData(page, c)
}

func (a *agentSession) ToolCall(c *gin.Context) {
	sessionID, ok := pathID(c)
	if !ok {
		return
	}
	call, err := agentSessionService.ToolCall(c.Request.Context(), utils.GetUserID(c), sessionID, c.Param("callId"))
	if err != nil {
		failAccount("获取工具调用详情失败", err, c)
		return
	}
	response.OkWithData(gin.H{"toolCall": call}, c)
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
	agentEventService.Forget(sessionID)
	response.OkWithMessage("删除成功", c)
}

func (a *agentSession) SendMessage(c *gin.Context) {
	sessionID, ok := pathID(c)
	if !ok {
		return
	}
	var input xyRequest.SendAgentMessage
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	headerKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if headerKey != "" && input.IdempotencyKey != "" && headerKey != strings.TrimSpace(input.IdempotencyKey) {
		response.FailWithMessage("幂等键不一致", c)
		return
	}
	if headerKey != "" {
		input.IdempotencyKey = headerKey
	}
	result, err := agentRunnerService.Submit(c.Request.Context(), utils.GetUserID(c), sessionID, input.Message, input.IdempotencyKey, input.Directives...)
	if err != nil {
		global.GVA_LOG.Error("提交 Agent 消息失败", zap.Error(err))
		response.FailWithMessage("消息提交失败，请检查 Agent 基础配置后重试", c)
		return
	}
	accepted := xyResponse.AgentRunAccepted{
		ID: result.Run.ID, SessionID: result.Run.SessionID, Status: result.Run.Status,
		StartedAt: result.Run.StartedAt, Assistant: result.Assistant,
	}
	response.OkWithData(gin.H{"run": accepted}, c)
}

func (a *agentSession) CancelRun(c *gin.Context) {
	sessionID, ok := pathID(c)
	if !ok {
		return
	}
	runID, err := strconv.ParseUint(c.Param("runId"), 10, 64)
	if err != nil || runID == 0 {
		response.FailWithMessage("运行 ID 无效", c)
		return
	}
	if err := agentRunnerService.Cancel(c.Request.Context(), utils.GetUserID(c), sessionID, uint(runID)); err != nil {
		response.FailWithMessage("取消 Agent 运行失败", c)
		return
	}
	response.OkWithMessage("已请求取消", c)
}

func (a *agentSession) ApproveRun(c *gin.Context) {
	runID, ok := pathID(c)
	if !ok {
		return
	}
	call, err := agentRunnerService.Approve(c.Request.Context(), utils.GetUserID(c), runID)
	if err != nil {
		global.GVA_LOG.Error("确认 Agent 操作失败", zap.Error(err))
		response.FailWithMessage("确认操作执行失败", c)
		return
	}
	response.OkWithData(gin.H{"tool": call.ToolName, "status": call.Status}, c)
}

func (a *agentSession) Events(c *gin.Context) {
	sessionID, ok := pathID(c)
	if !ok {
		return
	}
	if _, err := agentSessionService.Get(c.Request.Context(), utils.GetUserID(c), sessionID); err != nil {
		response.FailWithMessage("Agent 会话不存在", c)
		return
	}
	afterID, err := parseLastEventID(c.GetHeader("Last-Event-ID"))
	if err != nil {
		response.FailWithMessage("Last-Event-ID 无效", c)
		return
	}
	replay, events, unsubscribe, err := agentEventService.Subscribe(c.Request.Context(), sessionID, afterID)
	if err != nil {
		response.FailWithMessage("读取 Agent 事件失败", c)
		return
	}
	defer unsubscribe()
	header := c.Writer.Header()
	header.Set("Content-Type", "text/event-stream; charset=utf-8")
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("X-Accel-Buffering", "no")
	c.Status(200)
	if _, err := c.Writer.WriteString(": connected\n\n"); err != nil {
		return
	}
	c.Writer.Flush()
	for _, event := range replay {
		if err := writeAgentSSE(c, event.ID, event.Type, event.Data); err != nil {
			return
		}
	}
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case event, open := <-events:
			if !open {
				return
			}
			if err := writeAgentSSE(c, event.ID, event.Type, event.Data); err != nil {
				return
			}
		case <-heartbeat.C:
			if _, err := c.Writer.WriteString(": heartbeat\n\n"); err != nil {
				return
			}
			c.Writer.Flush()
		}
	}
}

func parseLastEventID(value string) (uint64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	return strconv.ParseUint(value, 10, 64)
}

func writeAgentSSE(c *gin.Context, id uint64, eventType string, data []byte) error {
	if id > 0 {
		if _, err := fmt.Fprintf(c.Writer, "id: %d\n", id); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", eventType, data); err != nil {
		return err
	}
	c.Writer.Flush()
	return nil
}
