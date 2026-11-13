package xianyu

import (
	"context"
	"github.com/gcc798/xianyu.bot/internal/model"

	"github.com/gcc798/xianyu.bot/internal/service"
	"gorm.io/gorm"
)

// Initialize 校验凭据密钥并启动所有已启用账号的消息连接。
func Initialize(ctx context.Context, db *gorm.DB, webSocket *service.WebSocketService) error {
	if err := service.ValidateCredentialKey(db); err != nil {
		return err
	}
	// 单实例启动时，上一进程已退出；只重试已持久化发送时间的确认操作。
	if err := db.WithContext(ctx).Model(&model.DeliveryTask{}).Where("status = ?", model.DeliveryTaskSending).
		Updates(map[string]any{"status": gorm.Expr("CASE WHEN sent_at IS NULL THEN ? ELSE ? END", model.DeliveryTaskUnknown, model.DeliveryTaskFailed), "last_error": "服务重启，发货任务中断"}).Error; err != nil {
		return err
	}
	webSocket.StartEnabledAccounts(ctx)
	return nil
}
