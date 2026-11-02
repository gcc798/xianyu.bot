package xianyu

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/initialize"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/service"
	"github.com/gin-gonic/gin"
)

func Initialize(engine *gin.Engine) error {
	ctx := context.Background()
	if err := initialize.Database(ctx); err != nil {
		return err
	}
	if err := service.ValidateCredentialKey(); err != nil {
		return err
	}
	initialize.Router(engine)
	service.Service.WebSocket.StartEnabledAccounts(ctx)
	return nil
}
