package router

import "github.com/gin-gonic/gin"

type telegramProxy struct{ *dependencies }

func (r *telegramProxy) Init(private *gin.RouterGroup) {
	private.Any("xy/proxy/telegram/*path", r.handlers.TelegramProxy.Proxy)
}
