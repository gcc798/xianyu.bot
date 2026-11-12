package request

// CreateAgentSession 描述新建 Agent 会话的输入。
type CreateAgentSession struct {
	Title string `json:"title" binding:"max=200"` // Title 是会话显示标题。
}

// UpdateAgentSession 描述会话名称修改输入。
type UpdateAgentSession struct {
	Title string `json:"title" binding:"required,max=200"` // Title 是新的会话显示标题。
}

// AgentSessionSearch 描述 Agent 会话列表分页条件。
type AgentSessionSearch struct {
	Pagination // Pagination 提供页码与每页数量。
}

// AgentTrajectorySearch 描述 Agent 事件轨迹的分页和筛选条件。
type AgentTrajectorySearch struct {
	Order    string `json:"order" form:"order" binding:"omitempty,oneof=asc desc"` // Order 指定事件时间顺序。
	CursorID uint64 `json:"cursorId" form:"cursorId"`                              // CursorID 是当前方向上次返回的最后事件 ID。
	From     string `json:"from" form:"from" binding:"omitempty,max=40"`           // From 是包含的 RFC3339 起始时间。
	To       string `json:"to" form:"to" binding:"omitempty,max=40"`               // To 是不包含的 RFC3339 结束时间。
	Limit    int    `json:"limit" form:"limit" binding:"omitempty,min=1,max=500"`  // Limit 限制单次返回的事件数量。
	RunID    uint   `json:"runId" form:"runId"`                                    // RunID 仅返回指定运行的事件。
	Type     string `json:"type" form:"type" binding:"omitempty,max=80"`           // Type 仅返回指定类别的事件。
}

// SendAgentMessage 描述启动一轮 Agent 对话的输入。
type SendAgentMessage struct {
	Message        string           `json:"message" binding:"required,max=20000"`       // Message 是用户发送给 Agent 的正文。
	IdempotencyKey string           `json:"idempotencyKey" binding:"omitempty,max=100"` // IdempotencyKey 防止客户端重复提交同一轮对话。
	Directives     []AgentDirective `json:"directives" binding:"max=2,dive"`            // Directives 限定本轮优先使用的技能或工具。
}

// AgentDirective 指定本轮 Agent 应关注的技能或工具。
type AgentDirective struct {
	Type string `json:"type" binding:"required,oneof=skill tool"` // Type 是 skill 或 tool 指令类型。
	Name string `json:"name" binding:"required,max=120"`          // Name 是技能或工具名称。
}
