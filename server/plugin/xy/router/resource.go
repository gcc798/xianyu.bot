package router

import "github.com/gin-gonic/gin"

type resource struct{}

func (r *resource) Init(private *gin.RouterGroup) {
	private.GET("xy/resources", resourceAPI.List)
	private.GET("xy/resources/:id", resourceAPI.Get)
	private.GET("xy/resource-tasks/:id", resourceAPI.GetTask)
}
