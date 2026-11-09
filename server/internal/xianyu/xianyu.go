package xianyu

import (
	"context"

	"github.com/gcc798/xianyu.bot/internal/service"
	"gorm.io/gorm"
)

// Initialize 校验凭据密钥并启动所有已启用账号的消息连接。
func Initialize(ctx context.Context, db *gorm.DB, webSocket *service.WebSocketService) error {
	if err := service.ValidateCredentialKey(db); err != nil {
		return err
	}
	webSocket.StartEnabledAccounts(ctx)
	return nil
}
