package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/router"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

func Router(engine *gin.Engine) {
	private := engine.Group("")
	private.Use(middleware.JWTAuth())
	internal := engine.Group("/xy/internal")
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
	router.Router.AgentSession.Init(private)
	router.Router.Resource.InitInternal(internal)
}
