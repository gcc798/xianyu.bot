package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

type TelegramProxyRouter struct{ handlers *api.API }

func NewTelegramProxyRouter(handlers *api.API) *TelegramProxyRouter {
	return &TelegramProxyRouter{handlers: handlers}
}

func (r *TelegramProxyRouter) Init(private *gin.RouterGroup) {
	private.Any("xy/proxy/telegram/*path", r.handlers.TelegramProxy.Proxy)
}
