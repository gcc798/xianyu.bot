package infra

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RunServer(router *gin.Engine, port int) {
	address := fmt.Sprintf(":%d", port)
	zap.L().Info("闲鱼 Harness Agent 启动", zap.String("address", address))
	initServer(address, router, 10*time.Minute, 10*time.Minute)
}
