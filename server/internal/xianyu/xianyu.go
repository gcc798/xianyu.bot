package xianyu

import (
	"context"
	"errors"

	"github.com/gcc798/go.xianyu.assistant/server/internal/api/middleware"
	xianyuRouter "github.com/gcc798/go.xianyu.assistant/server/internal/api/router/xianyu"
	"github.com/gcc798/go.xianyu.assistant/server/internal/api/xianyu"
	"github.com/gcc798/go.xianyu.assistant/server/internal/infra"
	"github.com/gcc798/go.xianyu.assistant/server/internal/model/xianyu"
	"github.com/gcc798/go.xianyu.assistant/server/internal/security"
	"github.com/gcc798/go.xianyu.assistant/server/internal/service/xianyu"
	"github.com/gin-gonic/gin"
)

func Initialize(engine *gin.Engine, jwt *security.JWT, deps service.Dependencies) error {
	ctx := context.Background()
	if err := initializeDatabase(ctx); err != nil {
		return err
	}
	if err := service.ValidateCredentialKey(); err != nil {
		return err
	}
	services := service.New(deps)
	handlers := api.New(services)
	private := engine.Group("")
	private.Use(middleware.JWTAuth(jwt))
	internal := engine.Group("/xy/internal")
	routes := xianyuRouter.New(handlers)
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
	routes.Resource.InitInternal(internal)
	services.WebSocket.StartEnabledAccounts(ctx)
	return nil
}

func initializeDatabase(ctx context.Context) error {
	if infra.DB == nil {
		return errors.New("database is not initialized")
	}
	config := model.DefaultAgentConfig()
	return infra.DB.WithContext(ctx).FirstOrCreate(&config, model.AgentConfig{ID: 1}).Error
}
