package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
)

const administratorAuthorityID = 888

type aiConfig struct{}

func (r *aiConfig) Init(private *gin.RouterGroup) {
	group := private.Group("xy/ai").Use(adminOnly())
	group.GET("models", aiConfigAPI.ListModels)
	group.POST("models", aiConfigAPI.CreateModel)
	group.PUT("models/:id", aiConfigAPI.UpdateModel)
	group.DELETE("models/:id", aiConfigAPI.DeleteModel)
	group.PUT("models/:id/default", aiConfigAPI.SetDefaultModel)
	group.GET("mcp-servers", aiConfigAPI.ListMCPServers)
	group.POST("mcp-servers", aiConfigAPI.CreateMCPServer)
	group.PUT("mcp-servers/:id", aiConfigAPI.UpdateMCPServer)
	group.DELETE("mcp-servers/:id", aiConfigAPI.DeleteMCPServer)
	group.GET("agent-config", aiConfigAPI.GetAgentConfig)
	group.PUT("agent-config", aiConfigAPI.UpdateAgentConfig)
}

func adminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if utils.GetUserAuthorityId(c) != administratorAuthorityID {
			response.NoAuth("仅管理员可管理 AI 基础配置", c)
			c.Abort()
			return
		}
		c.Next()
	}
}
