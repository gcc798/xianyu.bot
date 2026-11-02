package app

import (
	"net/http"

	"github.com/gcc798/xianyu.bot/internal/api/middleware"
	systemRouter "github.com/gcc798/xianyu.bot/internal/api/router/system"
	"github.com/gcc798/xianyu.bot/internal/security"
	xianyuService "github.com/gcc798/xianyu.bot/internal/service/xianyu"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Routers(systemRoutes *systemRouter.RouterGroup, jwt *security.JWT, xianyuDeps xianyuService.Dependencies) *gin.Engine {
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
	systemRoutes.InitSysDictionaryRouter(private)
	systemRoutes.InitSysDictionaryDetailRouter(private)
	systemRoutes.InitLoginLogRouter(private)
	initializeXianyu(engine, jwt, xianyuDeps)

	zap.L().Info("router register success")
	return engine
}
