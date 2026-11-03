package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

type BusinessEventRouter struct{ handlers *api.API }

func NewBusinessEventRouter(handlers *api.API) *BusinessEventRouter {
	return &BusinessEventRouter{handlers: handlers}
}

func (r *BusinessEventRouter) Init(group *gin.RouterGroup) {
	group.Group("xy/event").GET("list", r.handlers.BusinessEvent.List)
}
