package response

import (
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
