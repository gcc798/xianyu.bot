package router

import "github.com/gin-gonic/gin"

type businessEvent struct{}

func (r *businessEvent) Init(group *gin.RouterGroup) {
	group.Group("xy/event").GET("list", businessEventApi.List)
}
