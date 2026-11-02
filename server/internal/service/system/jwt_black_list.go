package system

import (
	"go.uber.org/zap"

	"github.com/gcc798/xianyu.bot/internal/auth"
	"github.com/gcc798/xianyu.bot/internal/model/system"
	"gorm.io/gorm"
)

type JwtService struct{ db *gorm.DB }

func (jwtService *JwtService) Revoke(token system.RevokedToken) error {
	if err := jwtService.db.Create(&token).Error; err != nil {
		return err
	}
	auth.RevokeToken(token.Token)
	return nil
}

func (jwtService *JwtService) LoadRevokedTokens() {
	var data []string
	err := jwtService.db.Model(&system.RevokedToken{}).Select("jwt").Find(&data).Error
	if err != nil {
		zap.L().Error("加载数据库jwt黑名单失败!", zap.Error(err))
		return
	}
	for _, token := range data {
		auth.RevokeToken(token)
	}
}
