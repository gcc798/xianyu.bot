package xianyu

import (
	"context"
	"errors"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/gcc798/xianyu.bot/internal/service"
	"gorm.io/gorm"
)

func Initialize(ctx context.Context, services *service.Service) error {
	if err := initializeDatabase(ctx, services.DB); err != nil {
		return err
	}
	if err := service.ValidateCredentialKey(services.DB); err != nil {
		return err
	}
	services.WebSocket.StartEnabledAccounts(ctx)
	return nil
}

func initializeDatabase(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return errors.New("database is not initialized")
	}
	config := model.DefaultAgentConfig()
	return db.WithContext(ctx).FirstOrCreate(&config, model.AgentConfig{ID: 1}).Error
}
