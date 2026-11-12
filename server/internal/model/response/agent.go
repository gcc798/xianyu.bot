package response

import (
	"encoding/json"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
)

// AgentSessionDetail 汇总会话、消息、工具轨迹、审批和资源任务。
type AgentSessionDetail struct {
	model.AgentSession                      // AgentSession 是会话基本信息。
	Messages           []model.AgentMessage `json:"messages"`         // Messages 是会话中已落定的消息。
	ToolTraces         []AgentToolTrace     `json:"toolTraces"`       // ToolTraces 是最近的工具调用轨迹。
	PendingApprovals   []AgentApproval      `json:"pendingApprovals"` // PendingApprovals 是等待用户确认的写工具调用。
	Tasks              []AgentTask          `json:"tasks"`            // Tasks 是会话关联的资源长任务。
}

// AgentToolTrace 是供会话详情展示的工具调用摘要。
type AgentToolTrace struct {
	ID             uint      `json:"id"`             // ID 是工具调用记录 ID。
	RunID          uint      `json:"runId"`          // RunID 是调用所属的 Agent 运行 ID。
	CallID         string    `json:"callId"`         // CallID 是模型生成的工具调用标识。
	ToolCode       string    `json:"toolCode"`       // ToolCode 是工具稳定业务编码。
	ToolName       string    `json:"toolName"`       // ToolName 是提供给模型的工具名称。
	ToolSource     string    `json:"toolSource"`     // ToolSource 标识 local 或 mcp 工具来源。
	Arguments      string    `json:"arguments"`      // Arguments 是脱敏后的工具入参。
	Result         string    `json:"result"`         // Result 是脱敏后的工具结果。
	Status         string    `json:"status"`         // Status 是工具调用状态。
	DurationMillis int64     `json:"durationMillis"` // DurationMillis 是调用耗时毫秒数。
	CreatedAt      time.Time `json:"createdAt"`      // CreatedAt 是调用记录创建时间。
}

// AgentTask 是供会话详情展示的资源任务进度。
type AgentTask struct {
	ID              uint   `json:"id"`              // ID 是资源任务 ID。
	RunID           uint   `json:"runId"`           // RunID 是发起任务的 Agent 运行 ID。
	State           string `json:"state"`           // State 是任务整体状态。
	Stage           string `json:"stage"`           // Stage 是当前处理阶段。
	ProgressPercent int    `json:"progressPercent"` // ProgressPercent 是零到一百的完成百分比。
	ProgressDone    int64  `json:"progressDone"`    // ProgressDone 是已完成工作量。
	ProgressTotal   int64  `json:"progressTotal"`   // ProgressTotal 是总工作量。
	ProgressUnit    string `json:"progressUnit"`    // ProgressUnit 是工作量单位。
}

// AgentApproval 描述一个等待用户确认的写工具调用。
type AgentApproval struct {
	RunID      uint      `json:"runId"`      // RunID 是审批所属的 Agent 运行 ID。
	ToolCode   string    `json:"toolCode"`   // ToolCode 是待执行工具的稳定业务编码。
	ToolName   string    `json:"toolName"`   // ToolName 是待执行工具的模型名称。
	ToolSource string    `json:"toolSource"` // ToolSource 标识 local 或 mcp 工具来源。
	CreatedAt  time.Time `json:"createdAt"`  // CreatedAt 是审批请求创建时间。
}

// AgentRunAccepted 表示 Agent 运行已创建或命中幂等结果。
type AgentRunAccepted struct {
	ID        uint                `json:"id"`                  // ID 是 Agent 运行 ID。
	SessionID uint                `json:"sessionId"`           // SessionID 是运行所属的会话 ID。
	Status    string              `json:"status"`              // Status 是运行当前状态。
	StartedAt time.Time           `json:"startedAt"`           // StartedAt 是运行开始时间。
	Assistant *model.AgentMessage `json:"assistant,omitempty"` // Assistant 是幂等命中时已有的助手消息。
}

// AgentTrajectoryEvent 是可增量读取的 Agent 审计事件。
type AgentTrajectoryEvent struct {
	ID           uint64          `json:"id"`                     // ID 是全局递增事件游标。
	RunID        *uint           `json:"runId,omitempty"`        // RunID 是事件所属的 Agent 运行 ID。
	TurnNo       *int            `json:"turnNo,omitempty"`       // TurnNo 是事件所在的对话轮次。
	StepNo       *int            `json:"stepNo,omitempty"`       // StepNo 是事件所在的运行步骤。
	Type         string          `json:"type"`                   // Type 是事件类别编码。
	CallID       string          `json:"callId,omitempty"`       // CallID 是关联的工具调用标识。
	ParentCallID string          `json:"parentCallId,omitempty"` // ParentCallID 是父级工具调用标识。
	Data         json.RawMessage `json:"data"`                   // Data 是事件类型对应的 JSON 负载。
	CreatedAt    time.Time       `json:"createdAt"`              // CreatedAt 是事件发生时间。
}

// AgentTrajectoryPage 描述一页 Agent 事件、分页游标和实时流位置。
type AgentTrajectoryPage struct {
	Events        []AgentTrajectoryEvent `json:"events"`        // Events 按请求的时间方向返回。
	NextCursorID  uint64                 `json:"nextCursorId"`  // NextCursorID 是下一页使用的边界事件 ID。
	LatestEventID uint64                 `json:"latestEventId"` // LatestEventID 是会话实时事件流的续接位置。
	HasMore       bool                   `json:"hasMore"`       // HasMore 表示当前条件下仍有下一页。
}

// AgentToolCallDetail 展示单次工具调用的完整审计信息。
type AgentToolCallDetail struct {
	ID             uint            `json:"id"`                    // ID 是工具调用记录 ID。
	RunID          uint            `json:"runId"`                 // RunID 是调用所属的 Agent 运行 ID。
	CallID         string          `json:"callId"`                // CallID 是模型生成的工具调用标识。
	MCPServerCode  string          `json:"mcpServerCode"`         // MCPServerCode 是远程工具所属的 MCP 服务编码。
	ToolCode       string          `json:"toolCode"`              // ToolCode 是工具稳定业务编码。
	ToolName       string          `json:"toolName"`              // ToolName 是提供给模型的工具名称。
	ToolSource     string          `json:"toolSource"`            // ToolSource 标识 local 或 mcp 工具来源。
	Arguments      json.RawMessage `json:"arguments"`             // Arguments 是脱敏后的 JSON 入参。
	Result         json.RawMessage `json:"result"`                // Result 是脱敏后的 JSON 结果。
	Status         string          `json:"status"`                // Status 是工具调用状态。
	DurationMillis int64           `json:"durationMillis"`        // DurationMillis 是调用耗时毫秒数。
	ErrorCode      string          `json:"errorCode"`             // ErrorCode 是稳定错误编码。
	ErrorMessage   string          `json:"errorMessage"`          // ErrorMessage 是调用失败时的安全错误描述。
	StartedAt      time.Time       `json:"startedAt"`             // StartedAt 是调用开始时间。
	CompletedAt    *time.Time      `json:"completedAt,omitempty"` // CompletedAt 是调用结束时间。
}
