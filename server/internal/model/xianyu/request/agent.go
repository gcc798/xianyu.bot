package request

import commonRequest "github.com/gcc798/xianyu.bot/internal/model/common/request"

type CreateAgentSession struct {
	Title string `json:"title" binding:"max=200"`
}

type AgentSessionSearch struct {
	commonRequest.Pagination
}

type AgentTrajectorySearch struct {
	AfterID uint64 `json:"afterId" form:"afterId"`
	Limit   int    `json:"limit" form:"limit" binding:"omitempty,min=1,max=500"`
	RunID   uint   `json:"runId" form:"runId"`
	Type    string `json:"type" form:"type" binding:"omitempty,max=80"`
}

type SendAgentMessage struct {
	Message        string           `json:"message" binding:"required,max=20000"`
	IdempotencyKey string           `json:"idempotencyKey" binding:"omitempty,max=100"`
	Directives     []AgentDirective `json:"directives" binding:"max=2,dive"`
}

type AgentDirective struct {
	Type string `json:"type" binding:"required,oneof=skill tool"`
	Name string `json:"name" binding:"required,max=120"`
}
