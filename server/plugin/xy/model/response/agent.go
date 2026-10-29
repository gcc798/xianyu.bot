package response

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model"
)

type AgentSessionDetail struct {
	model.AgentSession
	Messages         []model.AgentMessage `json:"messages"`
	PendingApprovals []AgentApproval      `json:"pendingApprovals"`
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
