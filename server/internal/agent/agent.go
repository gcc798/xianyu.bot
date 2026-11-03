package agent

import (
	"context"
	"errors"

	"github.com/gcc798/xianyu.bot/internal/model"
	"gorm.io/gorm"
)

func Initialize(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return errors.New("database is not initialized")
	}
	config := model.DefaultAgentConfig()
	return db.WithContext(ctx).FirstOrCreate(&config, model.AgentConfig{ID: 1}).Error
}

func rootContext(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
}
