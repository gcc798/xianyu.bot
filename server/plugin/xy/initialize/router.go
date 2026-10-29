package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/router"
	"github.com/gin-gonic/gin"
)

func Router(engine *gin.Engine) {
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix)
	private.Use(middleware.JWTAuth()).Use(middleware.CasbinHandler())
	router.Router.Status.Init(private)
	router.Router.Account.Init(private)
	router.Router.Product.Init(private)
	router.Router.ReplyRule.Init(private)
	router.Router.DeliveryRule.Init(private)
	router.Router.QRLogin.Init(private)
	router.Router.WebSocket.Init(private)
	router.Router.Message.Init(private)
	router.Router.Order.Init(private)
	router.Router.BusinessEvent.Init(private)
	router.Router.TelegramProxy.Init(private)
	router.Router.AIConfig.Init(private)
	router.Router.Resource.Init(private)
}
