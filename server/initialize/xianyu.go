package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func initializeXianyu(engine *gin.Engine) {
	start := func() {
		if err := xianyu.Initialize(engine); err != nil {
			global.GVA_LOG.Error("初始化闲鱼业务失败", zap.Error(err))
			panic(err)
		}
		global.GVA_ROUTERS = engine.Routes()
	}
	if global.GVA_DB == nil {
		GetDatabaseReadyNotifier().Subscribe(start)
		return
	}
	start()
}
