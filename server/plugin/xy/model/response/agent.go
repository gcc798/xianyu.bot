package response

import (
	"encoding/json"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model"
)

type AgentSessionDetail struct {
	model.AgentSession
	Messages         []model.AgentMessage `json:"messages"`
	ToolTraces       []AgentToolTrace     `json:"toolTraces"`
	PendingApprovals []AgentApproval      `json:"pendingApprovals"`
	Tasks            []AgentTask          `json:"tasks"`
}

type AgentToolTrace struct {
	ID             uint      `json:"id"`
	RunID          uint      `json:"runId"`
	ToolName       string    `json:"toolName"`
	Arguments      string    `json:"arguments"`
	Result         string    `json:"result"`
	Status         string    `json:"status"`
	DurationMillis int64     `json:"durationMillis"`
	CreatedAt      time.Time `json:"createdAt"`
}

type AgentTask struct {
	ID              uint   `json:"id"`
	RunID           uint   `json:"runId"`
	State           string `json:"state"`
	Stage           string `json:"stage"`
	ProgressPercent int    `json:"progressPercent"`
	ProgressDone    int64  `json:"progressDone"`
	ProgressTotal   int64  `json:"progressTotal"`
	ProgressUnit    string `json:"progressUnit"`
}

type AgentApproval struct {
	RunID     uint      `json:"runId"`
	ToolName  string    `json:"toolName"`
	CreatedAt time.Time `json:"createdAt"`
}

type AgentRunAccepted struct {
	ID        uint                `json:"id"`
	SessionID uint                `json:"sessionId"`
	Status    string              `json:"status"`
	StartedAt time.Time           `json:"startedAt"`
	Assistant *model.AgentMessage `json:"assistant,omitempty"`
}

type AgentTrajectoryEvent struct {
	ID           uint64          `json:"id"`
	RunID        *uint           `json:"runId,omitempty"`
	TurnNo       *int            `json:"turnNo,omitempty"`
	StepNo       *int            `json:"stepNo,omitempty"`
	Type         string          `json:"type"`
	CallID       string          `json:"callId,omitempty"`
	ParentCallID string          `json:"parentCallId,omitempty"`
	Data         json.RawMessage `json:"data"`
	CreatedAt    time.Time       `json:"createdAt"`
}

type AgentTrajectoryPage struct {
	Events      []AgentTrajectoryEvent `json:"events"`
	NextAfterID uint64                 `json:"nextAfterId"`
	HasMore     bool                   `json:"hasMore"`
}

type AgentToolCallDetail struct {
	ID             uint            `json:"id"`
	RunID          uint            `json:"runId"`
	CallID         string          `json:"callId"`
	MCPServerCode  string          `json:"mcpServerCode"`
	ToolName       string          `json:"toolName"`
	Arguments      json.RawMessage `json:"arguments"`
	Result         json.RawMessage `json:"result"`
	Status         string          `json:"status"`
	DurationMillis int64           `json:"durationMillis"`
	ErrorCode      string          `json:"errorCode"`
	ErrorMessage   string          `json:"errorMessage"`
	StartedAt      time.Time       `json:"startedAt"`
	CompletedAt    *time.Time      `json:"completedAt,omitempty"`
}
