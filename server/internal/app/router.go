package app

import (
	"context"
	"net/http"

	"github.com/gcc798/xianyu.bot/internal/api/middleware"
	systemRouter "github.com/gcc798/xianyu.bot/internal/api/router/system"
	"github.com/gcc798/xianyu.bot/internal/security"
	xianyuService "github.com/gcc798/xianyu.bot/internal/service/xianyu"
	"github.com/gcc798/xianyu.bot/internal/xianyu"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Routers(ctx context.Context, systemRoutes *systemRouter.RouterGroup, jwt *security.JWT, xianyuDeps xianyuService.Dependencies) *gin.Engine {
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

	systemRoutes.InitBaseRouter(public)
	systemRoutes.InitJwtRouter(private)
	systemRoutes.InitUserRouter(private)
	systemRoutes.DictionaryRouter.Init(private)
	systemRoutes.DictionaryDetailRouter.Init(private)
	systemRoutes.InitLoginLogRouter(private)
	if err := xianyu.Initialize(ctx, engine, jwt, xianyuDeps); err != nil {
		zap.L().Error("初始化闲鱼业务失败", zap.Error(err))
		panic(err)
	}

	zap.L().Info("router register success")
	return engine
}
