package router

import "github.com/gin-gonic/gin"

type businessEvent struct{ *dependencies }

func (r *businessEvent) Init(group *gin.RouterGroup) {
	group.Group("xy/event").GET("list", r.handlers.BusinessEvent.List)
}
