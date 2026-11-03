package request

// CreateAgentSession 描述新建 Agent 会话的输入。
type CreateAgentSession struct {
	// Title 是会话显示标题。
	Title string `json:"title" binding:"max=200"`
}

// AgentSessionSearch 描述 Agent 会话列表分页条件。
type AgentSessionSearch struct {
	// Pagination 提供页码与每页数量。
	Pagination
}

// AgentTrajectorySearch 描述 Agent 事件轨迹的增量查询条件。
type AgentTrajectorySearch struct {
	// AfterID 仅返回该事件 ID 之后的记录。
	AfterID uint64 `json:"afterId" form:"afterId"`
	// Limit 限制单次返回的事件数量。
	Limit int `json:"limit" form:"limit" binding:"omitempty,min=1,max=500"`
	// RunID 仅返回指定运行的事件。
	RunID uint `json:"runId" form:"runId"`
	// Type 仅返回指定类别的事件。
	Type string `json:"type" form:"type" binding:"omitempty,max=80"`
}

// SendAgentMessage 描述启动一轮 Agent 对话的输入。
type SendAgentMessage struct {
	// Message 是用户发送给 Agent 的正文。
	Message string `json:"message" binding:"required,max=20000"`
	// IdempotencyKey 防止客户端重复提交同一轮对话。
	IdempotencyKey string `json:"idempotencyKey" binding:"omitempty,max=100"`
	// Directives 限定本轮优先使用的技能或工具。
	Directives []AgentDirective `json:"directives" binding:"max=2,dive"`
}

// AgentDirective 指定本轮 Agent 应关注的技能或工具。
type AgentDirective struct {
	// Type 是 skill 或 tool 指令类型。
	Type string `json:"type" binding:"required,oneof=skill tool"`
	// Name 是技能或工具名称。
	Name string `json:"name" binding:"required,max=120"`
}
