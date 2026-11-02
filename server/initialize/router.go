package initialize

import (
	"net/http"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/flipped-aurora/gin-vue-admin/server/router"
	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {
	engine := gin.New()
	engine.Use(middleware.GinRecovery(true))
	if gin.Mode() == gin.DebugMode {
		engine.Use(gin.Logger())
	}

	public := engine.Group("")
	private := engine.Group("")
	private.Use(middleware.JWTAuth())

	public.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, "ok")
	})

	systemRouter := router.RouterGroupApp.System
	systemRouter.InitBaseRouter(public)
	systemRouter.InitJwtRouter(private)
	systemRouter.InitUserRouter(private)
	systemRouter.InitSysDictionaryRouter(private)
	systemRouter.InitSysDictionaryDetailRouter(private)
	systemRouter.InitLoginLogRouter(private)
	initializeXianyu(engine)

	global.GVA_LOG.Info("router register success")
	return engine
}
