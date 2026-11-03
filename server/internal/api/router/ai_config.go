package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gin-gonic/gin"
)

// administratorAuthorityID 是固定管理员角色标识。
const administratorAuthorityID = 888

// AIConfigRouter 注册Agent 配置 HTTP 路由。
type AIConfigRouter struct {
	handlers *api.API // handlers 提供 HTTP 处理器。
}

// NewAIConfigRouter 创建 AIConfigRouter。
func NewAIConfigRouter(handlers *api.API) *AIConfigRouter {
	return &AIConfigRouter{handlers: handlers}
}

// Init 注册当前模块路由。
func (r *AIConfigRouter) Init(private *gin.RouterGroup) {
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

// adminOnly 限制 AI 配置路由仅管理员访问。
func adminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if security.GetUserAuthorityID(c) != administratorAuthorityID {
			response.NoAuth("仅管理员可管理 AI 基础配置", c)
			c.Abort()
			return
		}
		c.Next()
	}
}
