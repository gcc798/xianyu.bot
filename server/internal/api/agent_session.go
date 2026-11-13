package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gcc798/xianyu.bot/internal/agent"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gcc798/xianyu.bot/internal/model/response"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AgentSessionAPI 处理Agent 会话 HTTP 请求。
type AgentSessionAPI struct {
	sessions *agent.Session  // sessions 管理 Agent 会话。
	runner   *agent.Runner   // runner 执行 Agent 运行。
	events   *agent.EventHub // events 提供 Agent 事件流。
}

// NewAgentSessionAPI 创建 AgentSessionAPI。
func NewAgentSessionAPI(sessions *agent.Session, runner *agent.Runner, events *agent.EventHub) *AgentSessionAPI {
	return &AgentSessionAPI{sessions: sessions, runner: runner, events: events}
}

// Create 处理创建请求。
func (a *AgentSessionAPI) Create(c *gin.Context) {
	var input xyRequest.CreateAgentSession
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	session, err := a.sessions.Create(c.Request.Context(), security.GetUserID(c), input)
	if err != nil {
		failAccount("创建 Agent 会话失败", err, c)
		return
	}
	response.OkWithData(gin.H{"session": session}, c)
}

// ToolCatalog 返回当前可用的 Agent 工具目录。
func (a *AgentSessionAPI) ToolCatalog(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	response.OkWithData(gin.H{"list": a.runner.ToolCatalog(c.Request.Context())}, c)
}

// List 处理列表查询请求。
func (a *AgentSessionAPI) List(c *gin.Context) {
	var input xyRequest.AgentSessionSearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	sessions, total, err := a.sessions.List(c.Request.Context(), security.GetUserID(c), input)
	if err != nil {
		failAccount("获取 Agent 会话失败", err, c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: sessions, Total: total, Page: input.Page, PageSize: input.PageSize}, "获取成功", c)
}

// Get 处理详情查询请求。
func (a *AgentSessionAPI) Get(c *gin.Context) {
	sessionID, ok := pathID(c)
	if !ok {
		return
	}
	session, err := a.sessions.Get(c.Request.Context(), security.GetUserID(c), sessionID)
	if err != nil {
		failAccount("获取 Agent 会话详情失败", err, c)
		return
	}
	response.OkWithData(gin.H{"session": session}, c)
}

// Update 处理会话名称修改请求。
func (a *AgentSessionAPI) Update(c *gin.Context) {
	sessionID, ok := pathID(c)
	if !ok {
		return
	}
	var input xyRequest.UpdateAgentSession
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := a.sessions.UpdateTitle(c.Request.Context(), security.GetUserID(c), sessionID, input); err != nil {
		failAccount("修改 Agent 会话名称失败", err, c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

// Trajectory 处理 Agent 运行轨迹查询请求。
func (a *AgentSessionAPI) Trajectory(c *gin.Context) {
	sessionID, ok := pathID(c)
	if !ok {
		return
	}
	var input xyRequest.AgentTrajectorySearch
	if err := c.ShouldBindQuery(&input); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	page, err := a.sessions.Trajectory(c.Request.Context(), security.GetUserID(c), sessionID, input)
	if err != nil {
		failAccount("获取 Agent 轨迹失败", err, c)
		return
	}
	response.OkWithData(page, c)
}

// ToolCall 处理 Agent 工具调用详情请求。
func (a *AgentSessionAPI) ToolCall(c *gin.Context) {
	sessionID, ok := pathID(c)
	if !ok {
		return
	}
	call, err := a.sessions.ToolCall(c.Request.Context(), security.GetUserID(c), sessionID, c.Param("callId"))
	if err != nil {
		failAccount("获取工具调用详情失败", err, c)
		return
	}
	response.OkWithData(gin.H{"toolCall": call}, c)
}

// Delete 处理删除请求。
func (a *AgentSessionAPI) Delete(c *gin.Context) {
	sessionID, ok := pathID(c)
	if !ok {
		return
	}
	if err := a.sessions.Delete(c.Request.Context(), security.GetUserID(c), sessionID); err != nil {
		failAccount("删除 Agent 会话失败", err, c)
		return
	}
	a.events.Forget(sessionID)
	response.OkWithMessage("删除成功", c)
}

// SendMessage 处理 Agent 消息提交请求。
func (a *AgentSessionAPI) SendMessage(c *gin.Context) {
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
	result, err := a.runner.Submit(c.Request.Context(), security.GetUserID(c), sessionID, input.Message, input.IdempotencyKey, input.Directives...)
	if err != nil {
		zap.L().Error("提交 Agent 消息失败", zap.Error(err))
		response.FailWithMessage("消息提交失败，请检查 Agent 基础配置后重试", c)
		return
	}
	accepted := xyResponse.AgentRunAccepted{
		ID: result.Run.ID, SessionID: result.Run.SessionID, Status: result.Run.Status,
		StartedAt: result.Run.StartedAt, Assistant: result.Assistant,
	}
	response.OkWithData(gin.H{"run": accepted}, c)
}

// CancelRun 处理 Agent 运行取消请求。
func (a *AgentSessionAPI) CancelRun(c *gin.Context) {
	sessionID, ok := pathID(c)
	if !ok {
		return
	}
	runID, err := strconv.ParseUint(c.Param("runId"), 10, 64)
	if err != nil || runID == 0 {
		response.FailWithMessage("运行 ID 无效", c)
		return
	}
	if err := a.runner.Cancel(c.Request.Context(), security.GetUserID(c), sessionID, uint(runID)); err != nil {
		response.FailWithMessage("取消 Agent 运行失败", c)
		return
	}
	response.OkWithMessage("已请求取消", c)
}

// ApproveRun 处理 Agent 工具审批请求。
func (a *AgentSessionAPI) ApproveRun(c *gin.Context) {
	runID, ok := pathID(c)
	if !ok {
		return
	}
	call, err := a.runner.Approve(c.Request.Context(), security.GetUserID(c), runID)
	if err != nil {
		zap.L().Error("确认 Agent 操作失败", zap.Error(err))
		response.FailWithMessage("确认操作执行失败", c)
		return
	}
	response.OkWithData(gin.H{"tool": call.ToolName, "status": call.Status}, c)
}

// RejectRun 处理 Agent 工具审批拒绝请求。
func (a *AgentSessionAPI) RejectRun(c *gin.Context) {
	runID, ok := pathID(c)
	if !ok {
		return
	}
	call, err := a.runner.Reject(c.Request.Context(), security.GetUserID(c), runID)
	if err != nil {
		zap.L().Error("拒绝 Agent 操作失败", zap.Error(err))
		response.FailWithMessage("拒绝操作失败", c)
		return
	}
	response.OkWithData(gin.H{"tool": call.ToolName, "status": call.Status}, c)
}

// Events 建立 Agent 事件流。
func (a *AgentSessionAPI) Events(c *gin.Context) {
	sessionID, ok := pathID(c)
	if !ok {
		return
	}
	if _, err := a.sessions.Get(c.Request.Context(), security.GetUserID(c), sessionID); err != nil {
		response.FailWithMessage("Agent 会话不存在", c)
		return
	}
	afterID, err := parseLastEventID(c.GetHeader("Last-Event-ID"))
	if err != nil {
		response.FailWithMessage("Last-Event-ID 无效", c)
		return
	}
	replay, events, unsubscribe, err := a.events.Subscribe(c.Request.Context(), sessionID, afterID)
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
	for {
		for _, event := range replay {
			if err := writeAgentSSE(c, event.ID, event.Type, event.Data); err != nil {
				return
			}
			afterID = event.ID
		}
		if len(replay) != agent.EventPageSize {
			break
		}
		replay, err = a.events.ReplayPage(c.Request.Context(), sessionID, afterID)
		if err != nil {
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
			if event.ID != 0 && event.ID <= afterID {
				continue
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

// parseLastEventID 解析 SSE 最后事件标识。
func parseLastEventID(value string) (uint64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	return strconv.ParseUint(value, 10, 64)
}

// writeAgentSSE 写入一条 Agent SSE 事件。
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
