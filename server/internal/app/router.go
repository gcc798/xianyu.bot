package app

import (
	"net/http"

	"github.com/gcc798/xianyu.bot/internal/api/middleware"
	"github.com/gcc798/xianyu.bot/internal/api/router"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Routers 创建并注册 HTTP 路由。
func Routers(routes *router.Router, jwt *security.JWT) *gin.Engine {
	engine := gin.New()
	engine.Use(middleware.GinRecovery(true))
	if gin.Mode() == gin.DebugMode {
		engine.Use(gin.Logger())
	}

	public := engine.Group("")
	private := engine.Group("")
	private.Use(middleware.JWTAuth(jwt))

	public.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, "ok")
	})

	routes.Auth.Init(public)
	routes.JWT.Init(private)
	routes.User.Init(private)
	routes.Dictionary.Init(private)
	routes.DictionaryDetail.Init(private)
	routes.LoginLog.Init(private)
	routes.Status.Init(private)
	routes.Account.Init(private)
	routes.Product.Init(private)
	routes.ReplyRule.Init(private)
	routes.DeliveryRule.Init(private)
	routes.QRLogin.Init(private)
	routes.WebSocket.Init(private)
	routes.Message.Init(private)
	routes.Order.Init(private)
	routes.BusinessEvent.Init(private)
	routes.TelegramProxy.Init(private)
	routes.AIConfig.Init(private)
	routes.Resource.Init(private)
	routes.AgentSession.Init(private)
	routes.Resource.InitInternal(engine.Group("/xy/internal"))

	zap.L().Info("router register success")
	return engine
}
