package service

import (
	"go.uber.org/zap"

	"github.com/gcc798/xianyu.bot/internal/auth"
	"github.com/gcc798/xianyu.bot/internal/model"
	"gorm.io/gorm"
)

// RevokedTokenService 持久化并恢复登录令牌撤销状态。
type RevokedTokenService struct {
	db *gorm.DB // db 保存跨重启生效的撤销令牌。
}

// NewRevokedTokenService 创建令牌撤销服务。
func NewRevokedTokenService(db *gorm.DB) *RevokedTokenService {
	return &RevokedTokenService{db: db}
}

// Revoke 先持久化 token，再将其加入当前进程撤销集合。
func (revokedTokens *RevokedTokenService) Revoke(token model.RevokedToken) error {
	if err := revokedTokens.db.Create(&token).Error; err != nil {
		return err
	}
	auth.RevokeToken(token.Token)
	return nil
}

// Load 在服务启动时把持久化黑名单恢复到进程内撤销集合。
func (revokedTokens *RevokedTokenService) Load() {
	var data []string
	err := revokedTokens.db.Model(&model.RevokedToken{}).Select("jwt").Find(&data).Error
	if err != nil {
		zap.L().Error("加载已撤销令牌失败", zap.Error(err))
		return
	}
	for _, token := range data {
		auth.RevokeToken(token)
	}
}
