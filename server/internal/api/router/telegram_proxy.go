package router

import (
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gin-gonic/gin"
)

// TelegramProxyRouter 注册Telegram 代理 HTTP 路由。
type TelegramProxyRouter struct {
	api *api.API // api 提供 HTTP 处理器。
}

// NewTelegramProxyRouter 创建 TelegramProxyRouter。
func NewTelegramProxyRouter(api *api.API) *TelegramProxyRouter {
	return &TelegramProxyRouter{api: api}
}

// Init 注册当前模块路由。
func (r *TelegramProxyRouter) Init(private *gin.RouterGroup) {
	private.Any("xy/proxy/telegram/*path", r.api.TelegramProxy.Proxy)
}
