package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// AgentSessionRouter 注册Agent 会话 HTTP 路由。
type AgentSessionRouter struct {
	handlers *api.API // handlers 提供 HTTP 处理器。
}

// NewAgentSessionRouter 创建 AgentSessionRouter。
func NewAgentSessionRouter(handlers *api.API) *AgentSessionRouter {
	return &AgentSessionRouter{handlers: handlers}
}

// Init 注册当前模块路由。
func (r *AgentSessionRouter) Init(private *gin.RouterGroup) {
	private.POST("xy/ai/agent/runs/:id/approve", r.handlers.AgentSession.ApproveRun)
	private.POST("xy/ai/agent/runs/:id/reject", r.handlers.AgentSession.RejectRun)
	private.GET("xy/ai/agent/tools", r.handlers.AgentSession.ToolCatalog)
	group := private.Group("xy/ai/agent/sessions")
	group.POST("", r.handlers.AgentSession.Create)
	group.GET("", r.handlers.AgentSession.List)
	group.GET(":id", r.handlers.AgentSession.Get)
	group.PUT(":id", r.handlers.AgentSession.Update)
	group.GET(":id/trajectory", r.handlers.AgentSession.Trajectory)
	group.GET(":id/tool-calls/:callId", r.handlers.AgentSession.ToolCall)
	group.DELETE(":id", r.handlers.AgentSession.Delete)
	group.POST(":id/messages", r.handlers.AgentSession.SendMessage)
	group.DELETE(":id/runs/:runId", r.handlers.AgentSession.CancelRun)
	group.GET(":id/events", r.handlers.AgentSession.Events)
}
