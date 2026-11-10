package router

import (
	"net/http"

	"github.com/gcc798/xianyu.bot/internal/api/middleware"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Engine 创建 Gin 引擎并注册公开、鉴权和内部路由。
func (r *Router) Engine(jwt *security.JWT) *gin.Engine {
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

	r.Auth.Init(public)
	r.User.Init(private)
	r.Dictionary.Init(private)
	r.DictionaryDetail.Init(private)
	r.LoginLog.Init(private)
	r.Status.Init(private)
	r.Account.Init(private)
	r.Product.Init(private)
	r.ReplyRule.Init(private)
	r.DeliveryRule.Init(private)
	r.QRLogin.Init(private)
	r.WebSocket.Init(private)
	r.Message.Init(private)
	r.Order.Init(private)
	r.BusinessEvent.Init(private)
	r.TelegramProxy.Init(private)
	r.AIConfig.Init(private)
	r.TelegramAPI.Init(private)
	r.Resource.Init(private)
	r.AgentSession.Init(private)
	r.Resource.InitInternal(engine.Group("/xy/internal"))
	r.TelegramAPI.InitInternal(engine.Group("/xy/internal"))

	zap.L().Info("router register success")
	return engine
}
