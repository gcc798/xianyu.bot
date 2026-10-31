package xianyu

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/initialize"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Initialize(engine *gin.Engine) error {
	ctx := context.Background()
	if err := initialize.Database(ctx); err != nil {
		return err
	}
	if err := service.ValidateCredentialKey(); err != nil {
		return err
	}
	if err := initialize.Naming(ctx); err != nil {
		return err
	}
	initialize.Api(ctx)
	initialize.Menu(ctx)
	if err := initialize.Authorization(ctx); err != nil {
		global.GVA_LOG.Error("初始化闲鱼助手管理员权限失败", zap.Error(err))
	}
	initialize.Router(engine)
	service.Service.WebSocket.StartEnabledAccounts(ctx)
	return nil
}
