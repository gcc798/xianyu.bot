package router

import "github.com/gin-gonic/gin"

type businessEvent struct{}

func (r *businessEvent) Init(group *gin.RouterGroup) {
	group.Group("xianyu/event").GET("list", businessEventApi.List)
}
