package core

import (
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/initialize"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"go.uber.org/zap"
)

func RunServer() {
	system.LoadAll()

	router := initialize.Routers()
	address := fmt.Sprintf(":%d", global.GVA_CONFIG.System.Addr)
	global.GVA_LOG.Info("闲鱼 Harness Agent 启动", zap.String("address", address))
	initServer(address, router, 10*time.Minute, 10*time.Minute)
}
