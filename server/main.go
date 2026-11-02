package main

import (
	"context"
	"os"

	"github.com/flipped-aurora/gin-vue-admin/server/core"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/initialize"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/database"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

func main() {
	initializeSystem()
	core.RunServer()
}

func initializeSystem() {
	core.Viper()
	if _, err := utils.ParseDuration(global.GVA_CONFIG.JWT.ExpiresTime); err != nil {
		panic("JWT expires-time 配置无效: " + err.Error())
	}
	if _, err := utils.ParseDuration(global.GVA_CONFIG.JWT.BufferTime); err != nil {
		panic("JWT buffer-time 配置无效: " + err.Error())
	}
	global.GVA_LOG = core.Zap()
	zap.ReplaceGlobals(global.GVA_LOG)
	global.GVA_DB = initialize.Postgres()
	if global.GVA_DB == nil {
		global.GVA_LOG.Error("PostgreSQL 配置缺少数据库名")
		os.Exit(1)
	}
	if err := database.Migrate(context.Background(), global.GVA_DB); err != nil {
		global.GVA_LOG.Error("数据库迁移失败", zap.Error(err))
		os.Exit(1)
	}
}
