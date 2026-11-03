package agent

import (
	"context"
	"errors"

	"github.com/gcc798/xianyu.bot/internal/model"
	"gorm.io/gorm"
)

// Initialize 校验数据库并创建唯一的默认 Agent 运行配置。
func Initialize(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return errors.New("database is not initialized")
	}
	config := model.DefaultAgentConfig()
	return db.WithContext(ctx).FirstOrCreate(&config, model.AgentConfig{ID: 1}).Error
}

// rootContext 保留有效根上下文，nil 时回退到后台上下文。
func rootContext(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
}
