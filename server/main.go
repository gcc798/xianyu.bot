package main

import (
	"context"
	"os"

	"github.com/flipped-aurora/gin-vue-admin/server/core"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/initialize"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/database"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

func main() {
	initializeSystem()
	core.RunServer()
}

func initializeSystem() {
	core.Viper()
	initialize.OtherInit()
	global.GVA_LOG = core.Zap()
	zap.ReplaceGlobals(global.GVA_LOG)
	global.GVA_DB = initialize.GormPgSql()
	if global.GVA_DB == nil {
		global.GVA_LOG.Error("PostgreSQL 配置缺少数据库名")
		os.Exit(1)
	}
	if err := database.Migrate(context.Background(), global.GVA_DB, global.GVA_CONFIG.System.DbType); err != nil {
		global.GVA_LOG.Error("数据库迁移失败", zap.Error(err))
		os.Exit(1)
	}
}
