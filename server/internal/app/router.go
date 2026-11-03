package app

import (
	"context"
	"net/http"

	"github.com/gcc798/xianyu.bot/internal/api/middleware"
	"github.com/gcc798/xianyu.bot/internal/api/router"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gcc798/xianyu.bot/internal/service"
	"github.com/gcc798/xianyu.bot/internal/xianyu"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Routers(ctx context.Context, routes *router.Router, jwt *security.JWT, services *service.Service) *gin.Engine {
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

	if err := xianyu.Initialize(ctx, services); err != nil {
		zap.L().Error("初始化闲鱼业务失败", zap.Error(err))
		panic(err)
	}

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
