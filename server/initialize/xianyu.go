package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func initializeXianyu(engine *gin.Engine) {
	if err := xianyu.Initialize(engine); err != nil {
		global.GVA_LOG.Error("初始化闲鱼业务失败", zap.Error(err))
		panic(err)
	}
}
