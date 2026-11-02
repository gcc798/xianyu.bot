package system

import (
	"go.uber.org/zap"

	"github.com/gcc798/xianyu.bot/internal/auth"
	"github.com/gcc798/xianyu.bot/internal/model/system"
	"gorm.io/gorm"
)

type RevokedTokenService struct{ db *gorm.DB }

func (revokedTokens *RevokedTokenService) Revoke(token system.RevokedToken) error {
	if err := revokedTokens.db.Create(&token).Error; err != nil {
		return err
	}
	auth.RevokeToken(token.Token)
	return nil
}

func (revokedTokens *RevokedTokenService) Load() {
	var data []string
	err := revokedTokens.db.Model(&system.RevokedToken{}).Select("jwt").Find(&data).Error
	if err != nil {
		zap.L().Error("加载已撤销令牌失败", zap.Error(err))
		return
	}
	for _, token := range data {
		auth.RevokeToken(token)
	}
}
