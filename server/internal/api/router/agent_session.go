package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

type AgentSessionRouter struct{ handlers *api.API }

func (r *AgentSessionRouter) Init(private *gin.RouterGroup) {
	private.POST("xy/ai/agent/runs/:id/approve", r.handlers.AgentSession.ApproveRun)
	group := private.Group("xy/ai/agent/sessions")
	group.POST("", r.handlers.AgentSession.Create)
	group.GET("", r.handlers.AgentSession.List)
	group.GET(":id", r.handlers.AgentSession.Get)
	group.GET(":id/trajectory", r.handlers.AgentSession.Trajectory)
	group.GET(":id/tool-calls/:callId", r.handlers.AgentSession.ToolCall)
	group.DELETE(":id", r.handlers.AgentSession.Delete)
	group.POST(":id/messages", r.handlers.AgentSession.SendMessage)
	group.DELETE(":id/runs/:runId", r.handlers.AgentSession.CancelRun)
	group.GET(":id/events", r.handlers.AgentSession.Events)
}
