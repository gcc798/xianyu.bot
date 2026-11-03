package model

import (
	"time"
)

// AgentSession 表示一个归属用户的连续 Agent 对话。
type AgentSession struct {
	// Timestamps 保存会话记录的生命周期字段。
	Timestamps
	// OwnerID 是会话所属的系统用户 ID。
	OwnerID uint `json:"-" gorm:"not null;index:idx_xy_agent_sessions_owner"`
	// Title 是会话显示标题。
	Title string `json:"title" gorm:"size:200;not null"`
	// Summary 是压缩后的历史上下文摘要。
	Summary string `json:"summary" gorm:"type:text"`
	// LastModelID 是最近一次使用的模型配置 ID。
	LastModelID *uint `json:"lastModelId"`
	// LastMessageAt 是最近一条消息时间。
	LastMessageAt *time.Time `json:"lastMessageAt" gorm:"index"`
}

// TableName 返回 Agent 会话表名。
func (AgentSession) TableName() string { return "xy_agent_sessions" }

// AgentMessage 保存一条已落定的会话消息。
type AgentMessage struct {
	// Timestamps 保存消息记录的生命周期字段。
	Timestamps
	// OwnerID 是消息所属的系统用户 ID。
	OwnerID uint `json:"-" gorm:"not null;index:idx_xy_agent_messages_owner"`
	// SessionID 是消息所属的 Agent 会话 ID。
	SessionID uint `json:"sessionId" gorm:"not null;index:idx_xy_agent_messages_session_sequence,unique"`
	// RunID 是生成消息的 Agent 运行 ID。
	RunID *uint `json:"runId" gorm:"index"`
	// Sequence 是会话内严格递增的消息序号。
	Sequence int64 `json:"sequence" gorm:"not null;index:idx_xy_agent_messages_session_sequence,unique"`
	// Role 是 user、assistant 或 system 消息角色。
	Role string `json:"role" gorm:"size:32;not null"`
	// Content 是消息正文。
	Content string `json:"content" gorm:"type:text;not null"`
	// Directives 是前端可执行指令的序列化值。
	Directives string `json:"directives" gorm:"type:text"`
}

// TableName 返回 Agent 消息表名。
func (AgentMessage) TableName() string { return "xy_agent_messages" }

// AgentRun 记录一次 Agent 推理执行及资源消耗。
type AgentRun struct {
	// Timestamps 保存运行记录的生命周期字段。
	Timestamps
	// OwnerID 是运行所属的系统用户 ID。
	OwnerID uint `json:"-" gorm:"not null;index:idx_xy_agent_runs_owner;uniqueIndex:idx_xy_agent_runs_owner_idempotency"`
	// SessionID 是运行所属的 Agent 会话 ID。
	SessionID uint `json:"sessionId" gorm:"not null;index"`
	// IdempotencyKey 防止同一用户重复创建运行。
	IdempotencyKey *string `json:"-" gorm:"size:100;uniqueIndex:idx_xy_agent_runs_owner_idempotency"`
	// ModelID 是本次使用的模型配置 ID。
	ModelID uint `json:"modelId" gorm:"not null"`
	// ModelName 是运行时模型名称快照。
	ModelName string `json:"modelName" gorm:"size:120;not null"`
	// Status 是运行中、成功、失败或取消状态。
	Status string `json:"status" gorm:"size:32;not null;default:running;index"`
	// EndReason 是运行停止原因编码。
	EndReason string `json:"endReason" gorm:"size:64"`
	// Iterations 是已完成的推理循环数。
	Iterations int `json:"iterations" gorm:"not null;default:0"`
	// ModelCalls 是模型调用总数。
	ModelCalls int `json:"modelCalls" gorm:"not null;default:0"`
	// ToolCalls 是工具调用总数。
	ToolCalls int `json:"toolCalls" gorm:"not null;default:0"`
	// InputTokens 是模型输入令牌总数。
	InputTokens int `json:"inputTokens" gorm:"not null;default:0"`
	// OutputTokens 是模型输出令牌总数。
	OutputTokens int `json:"outputTokens" gorm:"not null;default:0"`
	// WallTimeMillis 是运行墙钟耗时毫秒数。
	WallTimeMillis int64 `json:"wallTimeMillis" gorm:"not null;default:0"`
	// ConfigSnapshot 是本次运行配置的序列化快照。
	ConfigSnapshot string `json:"configSnapshot" gorm:"type:text;not null"`
	// ErrorMessage 是运行失败时的安全错误描述。
	ErrorMessage string `json:"errorMessage" gorm:"size:2000"`
	// StartedAt 是运行开始时间。
	StartedAt time.Time `json:"startedAt" gorm:"not null"`
	// CompletedAt 是运行结束时间。
	CompletedAt *time.Time `json:"completedAt"`
}

// TableName 返回 Agent 运行表名。
func (AgentRun) TableName() string { return "xy_agent_runs" }

// AgentRunTool 记录一次工具调用的身份、耗时和结果状态。
type AgentRunTool struct {
	// Timestamps 保存工具调用记录的生命周期字段。
	Timestamps
	// RunID 是调用所属的 Agent 运行 ID。
	RunID uint `json:"runId" gorm:"not null;index"`
	// CallID 是模型生成的工具调用唯一标识。
	CallID string `json:"callId" gorm:"size:200;not null;uniqueIndex"`
	// MCPServerCode 是远程工具所属的 MCP 服务编码。
	MCPServerCode string `json:"mcpServerCode" gorm:"size:64"`
	// ToolCode 是工具的稳定业务编码。
	ToolCode string `json:"toolCode" gorm:"size:120;not null"`
	// ToolName 是提供给模型的工具名称。
	ToolName string `json:"toolName" gorm:"size:120;not null"`
	// ToolSource 标识 local 或 mcp 工具来源。
	ToolSource string `json:"toolSource" gorm:"size:16;not null;default:mcp"`
	// Arguments 是工具入参快照，不向普通响应返回。
	Arguments string `json:"-" gorm:"type:text"`
	// Result 是工具结果快照，不向普通响应返回。
	Result string `json:"-" gorm:"type:text"`
	// Status 是调用执行状态。
	Status string `json:"status" gorm:"size:32;not null"`
	// DurationMillis 是调用耗时毫秒数。
	DurationMillis int64 `json:"durationMillis" gorm:"not null;default:0"`
	// ErrorCode 是稳定错误编码。
	ErrorCode string `json:"errorCode" gorm:"size:64"`
	// ErrorMessage 是调用失败时的安全错误描述。
	ErrorMessage string `json:"errorMessage" gorm:"size:1000"`
	// StartedAt 是调用开始时间。
	StartedAt time.Time `json:"startedAt" gorm:"not null"`
	// CompletedAt 是调用结束时间。
	CompletedAt *time.Time `json:"completedAt"`
}

// TableName 返回 Agent 工具调用表名。
func (AgentRunTool) TableName() string { return "xy_agent_run_tools" }

// AgentEvent 保存用于回放与审计会话的不可变事实，流式增量不落库。
type AgentEvent struct {
	// ID 是事件全局递增主键，也是流式恢复游标。
	ID uint64 `json:"id" gorm:"primaryKey;autoIncrement"`
	// OwnerID 是事件所属的系统用户 ID。
	OwnerID uint `json:"-" gorm:"not null;index:idx_xy_agent_events_owner_session,priority:1"`
	// SessionID 是事件所属的 Agent 会话 ID。
	SessionID uint `json:"sessionId" gorm:"not null;index:idx_xy_agent_events_owner_session,priority:2;index:idx_xy_agent_events_session_id"`
	// RunID 是事件所属的 Agent 运行 ID。
	RunID *uint `json:"runId,omitempty" gorm:"index:idx_xy_agent_events_run_id"`
	// TurnNo 是事件所在的对话轮次。
	TurnNo *int `json:"turnNo,omitempty"`
	// StepNo 是事件所在的运行步骤。
	StepNo *int `json:"stepNo,omitempty"`
	// EventType 是事件类别编码。
	EventType string `json:"type" gorm:"size:80;not null;index:idx_xy_agent_events_type"`
	// CallID 是关联的工具调用标识。
	CallID string `json:"callId,omitempty" gorm:"size:200;index:idx_xy_agent_events_call_id"`
	// ParentCallID 是父级工具调用标识。
	ParentCallID string `json:"parentCallId,omitempty" gorm:"size:200"`
	// Payload 是事件负载的序列化值，不向普通响应返回。
	Payload string `json:"-" gorm:"type:text;not null"`
	// CreatedAt 是事件发生时间。
	CreatedAt time.Time `json:"createdAt" gorm:"not null;index"`
}

// TableName 返回 Agent 事件表名。
func (AgentEvent) TableName() string { return "xy_agent_events" }
