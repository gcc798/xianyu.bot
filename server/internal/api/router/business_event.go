package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// BusinessEventRouter 注册业务事件 HTTP 路由。
type BusinessEventRouter struct {
	handlers *api.API // handlers 提供 HTTP 处理器。
}

// NewBusinessEventRouter 创建 BusinessEventRouter。
func NewBusinessEventRouter(handlers *api.API) *BusinessEventRouter {
	return &BusinessEventRouter{handlers: handlers}
}

// Init 注册当前模块路由。
func (r *BusinessEventRouter) Init(group *gin.RouterGroup) {
	group.Group("xy/event").GET("list", r.handlers.BusinessEvent.List)
}
