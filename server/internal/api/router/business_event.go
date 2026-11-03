package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

type BusinessEventRouter struct{ handlers *api.API }

func (r *BusinessEventRouter) Init(group *gin.RouterGroup) {
	group.Group("xy/event").GET("list", r.handlers.BusinessEvent.List)
}
