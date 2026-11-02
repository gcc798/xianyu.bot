package app

import (
	"net/http"

	"github.com/gcc798/go.xianyu.assistant/server/internal/api/middleware"
	systemRouter "github.com/gcc798/go.xianyu.assistant/server/internal/api/router/system"
	"github.com/gcc798/go.xianyu.assistant/server/internal/infra"
	"github.com/gcc798/go.xianyu.assistant/server/internal/security"
	xianyuService "github.com/gcc798/go.xianyu.assistant/server/internal/service/xianyu"
	"github.com/gin-gonic/gin"
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

	infra.Log.Info("router register success")
	return engine
}
