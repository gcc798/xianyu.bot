package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// AgentSessionRouter 注册Agent 会话 HTTP 路由。
type AgentSessionRouter struct {
	api *api.API // api 提供 HTTP 处理器。
}

// NewAgentSessionRouter 创建 AgentSessionRouter。
func NewAgentSessionRouter(api *api.API) *AgentSessionRouter {
	return &AgentSessionRouter{api: api}
}

// Init 注册当前模块路由。
func (r *AgentSessionRouter) Init(private *gin.RouterGroup) {
	private.POST("xy/ai/agent/runs/:id/approve", r.api.AgentSession.ApproveRun)
	private.POST("xy/ai/agent/runs/:id/reject", r.api.AgentSession.RejectRun)
	private.GET("xy/ai/agent/tools", r.api.AgentSession.ToolCatalog)
	group := private.Group("xy/ai/agent/sessions")
	group.POST("", r.api.AgentSession.Create)
	group.GET("", r.api.AgentSession.List)
	group.GET(":id", r.api.AgentSession.Get)
	group.PUT(":id", r.api.AgentSession.Update)
	group.GET(":id/trajectory", r.api.AgentSession.Trajectory)
	group.GET(":id/tool-calls/:callId", r.api.AgentSession.ToolCall)
	group.DELETE(":id", r.api.AgentSession.Delete)
	group.POST(":id/messages", r.api.AgentSession.SendMessage)
	group.DELETE(":id/runs/:runId", r.api.AgentSession.CancelRun)
	group.GET(":id/events", r.api.AgentSession.Events)
}
