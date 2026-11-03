package response

import (
	"encoding/json"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
)

// AgentSessionDetail 汇总会话、消息、工具轨迹、审批和资源任务。
type AgentSessionDetail struct {
	// AgentSession 是会话基本信息。
	model.AgentSession
	// Messages 是会话中已落定的消息。
	Messages []model.AgentMessage `json:"messages"`
	// ToolTraces 是最近的工具调用轨迹。
	ToolTraces []AgentToolTrace `json:"toolTraces"`
	// PendingApprovals 是等待用户确认的写工具调用。
	PendingApprovals []AgentApproval `json:"pendingApprovals"`
	// Tasks 是会话关联的资源长任务。
	Tasks []AgentTask `json:"tasks"`
}

// AgentToolTrace 是供会话详情展示的工具调用摘要。
type AgentToolTrace struct {
	// ID 是工具调用记录 ID。
	ID uint `json:"id"`
	// RunID 是调用所属的 Agent 运行 ID。
	RunID uint `json:"runId"`
	// CallID 是模型生成的工具调用标识。
	CallID string `json:"callId"`
	// ToolCode 是工具稳定业务编码。
	ToolCode string `json:"toolCode"`
	// ToolName 是提供给模型的工具名称。
	ToolName string `json:"toolName"`
	// ToolSource 标识 local 或 mcp 工具来源。
	ToolSource string `json:"toolSource"`
	// Arguments 是脱敏后的工具入参。
	Arguments string `json:"arguments"`
	// Result 是脱敏后的工具结果。
	Result string `json:"result"`
	// Status 是工具调用状态。
	Status string `json:"status"`
	// DurationMillis 是调用耗时毫秒数。
	DurationMillis int64 `json:"durationMillis"`
	// CreatedAt 是调用记录创建时间。
	CreatedAt time.Time `json:"createdAt"`
}

// AgentTask 是供会话详情展示的资源任务进度。
type AgentTask struct {
	// ID 是资源任务 ID。
	ID uint `json:"id"`
	// RunID 是发起任务的 Agent 运行 ID。
	RunID uint `json:"runId"`
	// State 是任务整体状态。
	State string `json:"state"`
	// Stage 是当前处理阶段。
	Stage string `json:"stage"`
	// ProgressPercent 是零到一百的完成百分比。
	ProgressPercent int `json:"progressPercent"`
	// ProgressDone 是已完成工作量。
	ProgressDone int64 `json:"progressDone"`
	// ProgressTotal 是总工作量。
	ProgressTotal int64 `json:"progressTotal"`
	// ProgressUnit 是工作量单位。
	ProgressUnit string `json:"progressUnit"`
}

// AgentApproval 描述一个等待用户确认的写工具调用。
type AgentApproval struct {
	// RunID 是审批所属的 Agent 运行 ID。
	RunID uint `json:"runId"`
	// ToolCode 是待执行工具的稳定业务编码。
	ToolCode string `json:"toolCode"`
	// ToolName 是待执行工具的模型名称。
	ToolName string `json:"toolName"`
	// ToolSource 标识 local 或 mcp 工具来源。
	ToolSource string `json:"toolSource"`
	// CreatedAt 是审批请求创建时间。
	CreatedAt time.Time `json:"createdAt"`
}

// AgentRunAccepted 表示 Agent 运行已创建或命中幂等结果。
type AgentRunAccepted struct {
	// ID 是 Agent 运行 ID。
	ID uint `json:"id"`
	// SessionID 是运行所属的会话 ID。
	SessionID uint `json:"sessionId"`
	// Status 是运行当前状态。
	Status string `json:"status"`
	// StartedAt 是运行开始时间。
	StartedAt time.Time `json:"startedAt"`
	// Assistant 是幂等命中时已有的助手消息。
	Assistant *model.AgentMessage `json:"assistant,omitempty"`
}

// AgentTrajectoryEvent 是可增量读取的 Agent 审计事件。
type AgentTrajectoryEvent struct {
	// ID 是全局递增事件游标。
	ID uint64 `json:"id"`
	// RunID 是事件所属的 Agent 运行 ID。
	RunID *uint `json:"runId,omitempty"`
	// TurnNo 是事件所在的对话轮次。
	TurnNo *int `json:"turnNo,omitempty"`
	// StepNo 是事件所在的运行步骤。
	StepNo *int `json:"stepNo,omitempty"`
	// Type 是事件类别编码。
	Type string `json:"type"`
	// CallID 是关联的工具调用标识。
	CallID string `json:"callId,omitempty"`
	// ParentCallID 是父级工具调用标识。
	ParentCallID string `json:"parentCallId,omitempty"`
	// Data 是事件类型对应的 JSON 负载。
	Data json.RawMessage `json:"data"`
	// CreatedAt 是事件发生时间。
	CreatedAt time.Time `json:"createdAt"`
}

// AgentTrajectoryPage 描述一页 Agent 事件及下次读取游标。
type AgentTrajectoryPage struct {
	// Events 是按 ID 升序返回的事件。
	Events []AgentTrajectoryEvent `json:"events"`
	// NextAfterID 是下一次增量查询应使用的游标。
	NextAfterID uint64 `json:"nextAfterId"`
	// HasMore 表示当前条件下仍有后续事件。
	HasMore bool `json:"hasMore"`
}

// AgentToolCallDetail 展示单次工具调用的完整审计信息。
type AgentToolCallDetail struct {
	// ID 是工具调用记录 ID。
	ID uint `json:"id"`
	// RunID 是调用所属的 Agent 运行 ID。
	RunID uint `json:"runId"`
	// CallID 是模型生成的工具调用标识。
	CallID string `json:"callId"`
	// MCPServerCode 是远程工具所属的 MCP 服务编码。
	MCPServerCode string `json:"mcpServerCode"`
	// ToolCode 是工具稳定业务编码。
	ToolCode string `json:"toolCode"`
	// ToolName 是提供给模型的工具名称。
	ToolName string `json:"toolName"`
	// ToolSource 标识 local 或 mcp 工具来源。
	ToolSource string `json:"toolSource"`
	// Arguments 是脱敏后的 JSON 入参。
	Arguments json.RawMessage `json:"arguments"`
	// Result 是脱敏后的 JSON 结果。
	Result json.RawMessage `json:"result"`
	// Status 是工具调用状态。
	Status string `json:"status"`
	// DurationMillis 是调用耗时毫秒数。
	DurationMillis int64 `json:"durationMillis"`
	// ErrorCode 是稳定错误编码。
	ErrorCode string `json:"errorCode"`
	// ErrorMessage 是调用失败时的安全错误描述。
	ErrorMessage string `json:"errorMessage"`
	// StartedAt 是调用开始时间。
	StartedAt time.Time `json:"startedAt"`
	// CompletedAt 是调用结束时间。
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}
