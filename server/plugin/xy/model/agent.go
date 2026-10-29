package model

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

type AgentSession struct {
	global.GVA_MODEL
	OwnerID       uint       `json:"-" gorm:"not null;index:idx_xy_agent_sessions_owner"`
	Title         string     `json:"title" gorm:"size:200;not null"`
	Summary       string     `json:"summary" gorm:"type:text"`
	LastModelID   *uint      `json:"lastModelId"`
	LastMessageAt *time.Time `json:"lastMessageAt" gorm:"index"`
}

func (AgentSession) TableName() string { return "xy_agent_sessions" }

type AgentMessage struct {
	global.GVA_MODEL
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
	global.GVA_MODEL
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
	global.GVA_MODEL
	RunID          uint       `json:"runId" gorm:"not null;index"`
	CallID         string     `json:"callId" gorm:"size:200;not null;uniqueIndex"`
	MCPServerCode  string     `json:"mcpServerCode" gorm:"size:64"`
	ToolName       string     `json:"toolName" gorm:"size:120;not null"`
	Arguments      string     `json:"-" gorm:"type:text"`
	Status         string     `json:"status" gorm:"size:32;not null"`
	DurationMillis int64      `json:"durationMillis" gorm:"not null;default:0"`
	ErrorCode      string     `json:"errorCode" gorm:"size:64"`
	ErrorMessage   string     `json:"errorMessage" gorm:"size:1000"`
	StartedAt      time.Time  `json:"startedAt" gorm:"not null"`
	CompletedAt    *time.Time `json:"completedAt"`
}

func (AgentRunTool) TableName() string { return "xy_agent_run_tools" }
