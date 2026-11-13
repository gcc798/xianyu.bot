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
	api *api.API // api 提供 HTTP 处理器。
}

// NewAIConfigRouter 创建 AIConfigRouter。
func NewAIConfigRouter(api *api.API) *AIConfigRouter {
	return &AIConfigRouter{api: api}
}

// Init 注册当前模块路由。
func (r *AIConfigRouter) Init(private *gin.RouterGroup) {
	group := private.Group("xy/ai").Use(adminOnly())
	group.GET("models", r.api.AIConfig.ListModels)
	group.POST("models", r.api.AIConfig.CreateModel)
	group.PUT("models/:id", r.api.AIConfig.UpdateModel)
	group.DELETE("models/:id", r.api.AIConfig.DeleteModel)
	group.GET("mcp-servers", r.api.AIConfig.ListMCPServers)
	group.POST("mcp-servers", r.api.AIConfig.CreateMCPServer)
	group.PUT("mcp-servers/:id", r.api.AIConfig.UpdateMCPServer)
	group.DELETE("mcp-servers/:id", r.api.AIConfig.DeleteMCPServer)
	group.GET("agent-config", r.api.AIConfig.GetAgentConfig)
	group.PUT("agent-config", r.api.AIConfig.UpdateAgentConfig)
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
