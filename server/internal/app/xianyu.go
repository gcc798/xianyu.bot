package app

import (
	"github.com/gcc798/go.xianyu.assistant/server/internal/infra"
	"github.com/gcc798/go.xianyu.assistant/server/internal/security"
	xianyuService "github.com/gcc798/go.xianyu.assistant/server/internal/service/xianyu"
	"github.com/gcc798/go.xianyu.assistant/server/internal/xianyu"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func initializeXianyu(engine *gin.Engine, jwt *security.JWT, deps xianyuService.Dependencies) {
	if err := xianyu.Initialize(engine, jwt, deps); err != nil {
		infra.Log.Error("初始化闲鱼业务失败", zap.Error(err))
		panic(err)
	}
}
