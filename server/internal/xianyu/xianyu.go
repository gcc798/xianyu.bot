package xianyu

import (
	"context"

	"github.com/gcc798/xianyu.bot/internal/service"
)

func Initialize(ctx context.Context, services *service.Service) error {
	if err := service.ValidateCredentialKey(services.DB); err != nil {
		return err
	}
	services.WebSocket.StartEnabledAccounts(ctx)
	return nil
}
