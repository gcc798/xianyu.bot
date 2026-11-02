package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RunServer(ctx context.Context, router *gin.Engine, port int) error {
	address := fmt.Sprintf(":%d", port)
	zap.L().Info("闲鱼 Harness Agent 启动", zap.String("address", address))
	return initServer(ctx, address, router, 10*time.Minute, 10*time.Minute)
}
