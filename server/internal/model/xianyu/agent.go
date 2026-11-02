package model

import (
	"time"

	"github.com/gcc798/go.xianyu.assistant/server/internal/model/common"
)

type AgentSession struct {
	common.Timestamps
	OwnerID       uint       `json:"-" gorm:"not null;index:idx_xy_agent_sessions_owner"`
	Title         string     `json:"title" gorm:"size:200;not null"`
	Summary       string     `json:"summary" gorm:"type:text"`
	LastModelID   *uint      `json:"lastModelId"`
	LastMessageAt *time.Time `json:"lastMessageAt" gorm:"index"`
}

func (AgentSession) TableName() string { return "xy_agent_sessions" }

type AgentMessage struct {
	common.Timestamps
	OwnerID    uint   `json:"-" gorm:"not null;index:idx_xy_agent_messages_owner"`
	SessionID  uint   `json:"sessionId" gorm:"not null;index:idx_xy_agent_messages_session_sequence,unique"`
	RunID      *uint  `json:"runId" gorm:"index"`
	Sequence   int64  `json:"sequence" gorm:"not null;index:idx_xy_agent_messages_session_sequence,unique"`
	Role       string `json:"role" gorm:"size:32;not null"`
	Content    string `json:"content" gorm:"type:text;not null"`
	Directives string `json:"directives" gorm:"type:text"`
}

func (AgentMessage) TableName() string { return "xy_agent_messages" }

type AgentRun struct {
	common.Timestamps
	OwnerID        uint       `json:"-" gorm:"not null;index:idx_xy_agent_runs_owner;uniqueIndex:idx_xy_agent_runs_owner_idempotency"`
	SessionID      uint       `json:"sessionId" gorm:"not null;index"`
	IdempotencyKey *string    `json:"-" gorm:"size:100;uniqueIndex:idx_xy_agent_runs_owner_idempotency"`
	ModelID        uint       `json:"modelId" gorm:"not null"`
	ModelName      string     `json:"modelName" gorm:"size:120;not null"`
	Status         string     `json:"status" gorm:"size:32;not null;default:running;index"`
	EndReason      string     `json:"endReason" gorm:"size:64"`
	Iterations     int        `json:"iterations" gorm:"not null;default:0"`
	ModelCalls     int        `json:"modelCalls" gorm:"not null;default:0"`
	ToolCalls      int        `json:"toolCalls" gorm:"not null;default:0"`
	InputTokens    int        `json:"inputTokens" gorm:"not null;default:0"`
	OutputTokens   int        `json:"outputTokens" gorm:"not null;default:0"`
	WallTimeMillis int64      `json:"wallTimeMillis" gorm:"not null;default:0"`
	ConfigSnapshot string     `json:"configSnapshot" gorm:"type:text;not null"`
	ErrorMessage   string     `json:"errorMessage" gorm:"size:2000"`
	StartedAt      time.Time  `json:"startedAt" gorm:"not null"`
	CompletedAt    *time.Time `json:"completedAt"`
}

func (AgentRun) TableName() string { return "xy_agent_runs" }

type AgentRunTool struct {
	common.Timestamps
	RunID          uint       `json:"runId" gorm:"not null;index"`
	CallID         string     `json:"callId" gorm:"size:200;not null;uniqueIndex"`
	MCPServerCode  string     `json:"mcpServerCode" gorm:"size:64"`
	ToolName       string     `json:"toolName" gorm:"size:120;not null"`
	Arguments      string     `json:"-" gorm:"type:text"`
	Result         string     `json:"-" gorm:"type:text"`
	Status         string     `json:"status" gorm:"size:32;not null"`
	DurationMillis int64      `json:"durationMillis" gorm:"not null;default:0"`
	ErrorCode      string     `json:"errorCode" gorm:"size:64"`
	ErrorMessage   string     `json:"errorMessage" gorm:"size:1000"`
	StartedAt      time.Time  `json:"startedAt" gorm:"not null"`
	CompletedAt    *time.Time `json:"completedAt"`
}

func (AgentRunTool) TableName() string { return "xy_agent_run_tools" }

// AgentEvent records the immutable facts used to replay and audit one Agent session.
// Streaming deltas stay transient; settled messages and execution boundaries are persisted here.
type AgentEvent struct {
	ID           uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	OwnerID      uint      `json:"-" gorm:"not null;index:idx_xy_agent_events_owner_session,priority:1"`
	SessionID    uint      `json:"sessionId" gorm:"not null;index:idx_xy_agent_events_owner_session,priority:2;index:idx_xy_agent_events_session_id"`
	RunID        *uint     `json:"runId,omitempty" gorm:"index:idx_xy_agent_events_run_id"`
	TurnNo       *int      `json:"turnNo,omitempty"`
	StepNo       *int      `json:"stepNo,omitempty"`
	EventType    string    `json:"type" gorm:"size:80;not null;index:idx_xy_agent_events_type"`
	CallID       string    `json:"callId,omitempty" gorm:"size:200;index:idx_xy_agent_events_call_id"`
	ParentCallID string    `json:"parentCallId,omitempty" gorm:"size:200"`
	Payload      string    `json:"-" gorm:"type:text;not null"`
	CreatedAt    time.Time `json:"createdAt" gorm:"not null;index"`
}

func (AgentEvent) TableName() string { return "xy_agent_events" }
