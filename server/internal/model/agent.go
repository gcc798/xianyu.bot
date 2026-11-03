package model

import (
	"time"
)

// AgentSession 表示一个归属用户的连续 Agent 对话。
type AgentSession struct {
	Timestamps               // Timestamps 保存会话记录的生命周期字段。
	OwnerID       uint       `json:"-" gorm:"not null;index:idx_xy_agent_sessions_owner"` // OwnerID 是会话所属的系统用户 ID。
	Title         string     `json:"title" gorm:"size:200;not null"`                      // Title 是会话显示标题。
	Summary       string     `json:"summary" gorm:"type:text"`                            // Summary 是压缩后的历史上下文摘要。
	LastModelID   *uint      `json:"lastModelId"`                                         // LastModelID 是最近一次使用的模型配置 ID。
	LastMessageAt *time.Time `json:"lastMessageAt" gorm:"index"`                          // LastMessageAt 是最近一条消息时间。
}

// TableName 返回 Agent 会话表名。
func (AgentSession) TableName() string { return "xy_agent_sessions" }

// AgentMessage 保存一条已落定的会话消息。
type AgentMessage struct {
	Timestamps        // Timestamps 保存消息记录的生命周期字段。
	OwnerID    uint   `json:"-" gorm:"not null;index:idx_xy_agent_messages_owner"`                           // OwnerID 是消息所属的系统用户 ID。
	SessionID  uint   `json:"sessionId" gorm:"not null;index:idx_xy_agent_messages_session_sequence,unique"` // SessionID 是消息所属的 Agent 会话 ID。
	RunID      *uint  `json:"runId" gorm:"index"`                                                            // RunID 是生成消息的 Agent 运行 ID。
	Sequence   int64  `json:"sequence" gorm:"not null;index:idx_xy_agent_messages_session_sequence,unique"`  // Sequence 是会话内严格递增的消息序号。
	Role       string `json:"role" gorm:"size:32;not null"`                                                  // Role 是 user、assistant 或 system 消息角色。
	Content    string `json:"content" gorm:"type:text;not null"`                                             // Content 是消息正文。
	Directives string `json:"directives" gorm:"type:text"`                                                   // Directives 是前端可执行指令的序列化值。
}

// TableName 返回 Agent 消息表名。
func (AgentMessage) TableName() string { return "xy_agent_messages" }

// AgentRun 记录一次 Agent 推理执行及资源消耗。
type AgentRun struct {
	Timestamps                // Timestamps 保存运行记录的生命周期字段。
	OwnerID        uint       `json:"-" gorm:"not null;index:idx_xy_agent_runs_owner;uniqueIndex:idx_xy_agent_runs_owner_idempotency"` // OwnerID 是运行所属的系统用户 ID。
	SessionID      uint       `json:"sessionId" gorm:"not null;index"`                                                                 // SessionID 是运行所属的 Agent 会话 ID。
	IdempotencyKey *string    `json:"-" gorm:"size:100;uniqueIndex:idx_xy_agent_runs_owner_idempotency"`                               // IdempotencyKey 防止同一用户重复创建运行。
	ModelID        uint       `json:"modelId" gorm:"not null"`                                                                         // ModelID 是本次使用的模型配置 ID。
	ModelName      string     `json:"modelName" gorm:"size:120;not null"`                                                              // ModelName 是运行时模型名称快照。
	Status         string     `json:"status" gorm:"size:32;not null;default:running;index"`                                            // Status 是运行中、成功、失败或取消状态。
	EndReason      string     `json:"endReason" gorm:"size:64"`                                                                        // EndReason 是运行停止原因编码。
	Iterations     int        `json:"iterations" gorm:"not null;default:0"`                                                            // Iterations 是已完成的推理循环数。
	ModelCalls     int        `json:"modelCalls" gorm:"not null;default:0"`                                                            // ModelCalls 是模型调用总数。
	ToolCalls      int        `json:"toolCalls" gorm:"not null;default:0"`                                                             // ToolCalls 是工具调用总数。
	InputTokens    int        `json:"inputTokens" gorm:"not null;default:0"`                                                           // InputTokens 是模型输入令牌总数。
	OutputTokens   int        `json:"outputTokens" gorm:"not null;default:0"`                                                          // OutputTokens 是模型输出令牌总数。
	WallTimeMillis int64      `json:"wallTimeMillis" gorm:"not null;default:0"`                                                        // WallTimeMillis 是运行墙钟耗时毫秒数。
	ConfigSnapshot string     `json:"configSnapshot" gorm:"type:text;not null"`                                                        // ConfigSnapshot 是本次运行配置的序列化快照。
	ErrorMessage   string     `json:"errorMessage" gorm:"size:2000"`                                                                   // ErrorMessage 是运行失败时的安全错误描述。
	StartedAt      time.Time  `json:"startedAt" gorm:"not null"`                                                                       // StartedAt 是运行开始时间。
	CompletedAt    *time.Time `json:"completedAt"`                                                                                     // CompletedAt 是运行结束时间。
}

// TableName 返回 Agent 运行表名。
func (AgentRun) TableName() string { return "xy_agent_runs" }

// AgentRunTool 记录一次工具调用的身份、耗时和结果状态。
type AgentRunTool struct {
	Timestamps                // Timestamps 保存工具调用记录的生命周期字段。
	RunID          uint       `json:"runId" gorm:"not null;index"`                    // RunID 是调用所属的 Agent 运行 ID。
	CallID         string     `json:"callId" gorm:"size:200;not null;uniqueIndex"`    // CallID 是模型生成的工具调用唯一标识。
	MCPServerCode  string     `json:"mcpServerCode" gorm:"size:64"`                   // MCPServerCode 是远程工具所属的 MCP 服务编码。
	ToolCode       string     `json:"toolCode" gorm:"size:120;not null"`              // ToolCode 是工具的稳定业务编码。
	ToolName       string     `json:"toolName" gorm:"size:120;not null"`              // ToolName 是提供给模型的工具名称。
	ToolSource     string     `json:"toolSource" gorm:"size:16;not null;default:mcp"` // ToolSource 标识 local 或 mcp 工具来源。
	Arguments      string     `json:"-" gorm:"type:text"`                             // Arguments 是工具入参快照，不向普通响应返回。
	Result         string     `json:"-" gorm:"type:text"`                             // Result 是工具结果快照，不向普通响应返回。
	Status         string     `json:"status" gorm:"size:32;not null"`                 // Status 是调用执行状态。
	DurationMillis int64      `json:"durationMillis" gorm:"not null;default:0"`       // DurationMillis 是调用耗时毫秒数。
	ErrorCode      string     `json:"errorCode" gorm:"size:64"`                       // ErrorCode 是稳定错误编码。
	ErrorMessage   string     `json:"errorMessage" gorm:"size:1000"`                  // ErrorMessage 是调用失败时的安全错误描述。
	StartedAt      time.Time  `json:"startedAt" gorm:"not null"`                      // StartedAt 是调用开始时间。
	CompletedAt    *time.Time `json:"completedAt"`                                    // CompletedAt 是调用结束时间。
}

// TableName 返回 Agent 工具调用表名。
func (AgentRunTool) TableName() string { return "xy_agent_run_tools" }

// AgentEvent 保存用于回放与审计会话的不可变事实，流式增量不落库。
type AgentEvent struct {
	ID           uint64    `json:"id" gorm:"primaryKey;autoIncrement"`                                                                                // ID 是事件全局递增主键，也是流式恢复游标。
	OwnerID      uint      `json:"-" gorm:"not null;index:idx_xy_agent_events_owner_session,priority:1"`                                              // OwnerID 是事件所属的系统用户 ID。
	SessionID    uint      `json:"sessionId" gorm:"not null;index:idx_xy_agent_events_owner_session,priority:2;index:idx_xy_agent_events_session_id"` // SessionID 是事件所属的 Agent 会话 ID。
	RunID        *uint     `json:"runId,omitempty" gorm:"index:idx_xy_agent_events_run_id"`                                                           // RunID 是事件所属的 Agent 运行 ID。
	TurnNo       *int      `json:"turnNo,omitempty"`                                                                                                  // TurnNo 是事件所在的对话轮次。
	StepNo       *int      `json:"stepNo,omitempty"`                                                                                                  // StepNo 是事件所在的运行步骤。
	EventType    string    `json:"type" gorm:"size:80;not null;index:idx_xy_agent_events_type"`                                                       // EventType 是事件类别编码。
	CallID       string    `json:"callId,omitempty" gorm:"size:200;index:idx_xy_agent_events_call_id"`                                                // CallID 是关联的工具调用标识。
	ParentCallID string    `json:"parentCallId,omitempty" gorm:"size:200"`                                                                            // ParentCallID 是父级工具调用标识。
	Payload      string    `json:"-" gorm:"type:text;not null"`                                                                                       // Payload 是事件负载的序列化值，不向普通响应返回。
	CreatedAt    time.Time `json:"createdAt" gorm:"not null;index"`                                                                                   // CreatedAt 是事件发生时间。
}

// TableName 返回 Agent 事件表名。
func (AgentEvent) TableName() string { return "xy_agent_events" }
