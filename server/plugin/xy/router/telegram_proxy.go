package router

import "github.com/gin-gonic/gin"

type telegramProxy struct{}

func (r *telegramProxy) Init(private *gin.RouterGroup) {
	private.Any("xy/proxy/telegram/*path", telegramProxyApi.Proxy)
}
