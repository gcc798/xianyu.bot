package xianyu

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xianyu/initialize"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xianyu/service"
	interfaces "github.com/flipped-aurora/gin-vue-admin/server/utils/plugin/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var _ interfaces.Plugin = (*plugin)(nil)

var Plugin = new(plugin)

type plugin struct{}

func init() {
	interfaces.Register(Plugin)
}

func (p *plugin) Register(engine *gin.Engine) {
	ctx := context.Background()
	if err := initialize.Database(ctx); err != nil {
		global.GVA_LOG.Error("初始化闲鱼助手 SQLite 配置失败", zap.Error(err))
	}
	initialize.Api(ctx)
	initialize.Menu(ctx)
	if err := initialize.Authorization(ctx); err != nil {
		global.GVA_LOG.Error("初始化闲鱼助手管理员权限失败", zap.Error(err))
	}
	initialize.Router(engine)
	service.Service.WebSocket.StartEnabledAccounts(ctx)
}
