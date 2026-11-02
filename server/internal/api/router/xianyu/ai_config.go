package router

import (
	"github.com/gcc798/xianyu.bot/internal/model/common/response"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gin-gonic/gin"
)

const administratorAuthorityID = 888

type aiConfig struct{ *dependencies }

func (r *aiConfig) Init(private *gin.RouterGroup) {
	group := private.Group("xy/ai").Use(adminOnly())
	group.GET("models", r.handlers.AIConfig.ListModels)
	group.POST("models", r.handlers.AIConfig.CreateModel)
	group.PUT("models/:id", r.handlers.AIConfig.UpdateModel)
	group.DELETE("models/:id", r.handlers.AIConfig.DeleteModel)
	group.PUT("models/:id/default", r.handlers.AIConfig.SetDefaultModel)
	group.GET("mcp-servers", r.handlers.AIConfig.ListMCPServers)
	group.POST("mcp-servers", r.handlers.AIConfig.CreateMCPServer)
	group.PUT("mcp-servers/:id", r.handlers.AIConfig.UpdateMCPServer)
	group.DELETE("mcp-servers/:id", r.handlers.AIConfig.DeleteMCPServer)
	group.GET("agent-config", r.handlers.AIConfig.GetAgentConfig)
	group.PUT("agent-config", r.handlers.AIConfig.UpdateAgentConfig)
}

func adminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if security.GetUserAuthorityId(c) != administratorAuthorityID {
			response.NoAuth("仅管理员可管理 AI 基础配置", c)
			c.Abort()
			return
		}
		c.Next()
	}
}
