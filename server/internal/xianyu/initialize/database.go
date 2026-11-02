package initialize

import (
	"context"
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model"
)

func Database(ctx context.Context) error {
	if global.GVA_DB == nil {
		return errors.New("database is not initialized")
	}

	config := model.DefaultAgentConfig()
	return global.GVA_DB.WithContext(ctx).FirstOrCreate(&config, model.AgentConfig{ID: 1}).Error
}
