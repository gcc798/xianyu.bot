package router

import "github.com/gin-gonic/gin"

type agentSession struct{}

func (r *agentSession) Init(private *gin.RouterGroup) {
	group := private.Group("xy/ai/agent/sessions")
	group.POST("", agentSessionAPI.Create)
	group.GET("", agentSessionAPI.List)
	group.GET(":id", agentSessionAPI.Get)
	group.DELETE(":id", agentSessionAPI.Delete)
	group.POST(":id/messages", agentSessionAPI.SendMessage)
	group.DELETE(":id/runs/:runId", agentSessionAPI.CancelRun)
	group.GET(":id/events", agentSessionAPI.Events)
}
