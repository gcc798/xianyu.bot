package system

import (
	"go.uber.org/zap"

	"github.com/gcc798/go.xianyu.assistant/server/internal/auth"
	"github.com/gcc798/go.xianyu.assistant/server/internal/infra"
	"github.com/gcc798/go.xianyu.assistant/server/internal/model/system"
	"gorm.io/gorm"
)

type JwtService struct{ db *gorm.DB }

//@author: [piexlmax](https://github.com/piexlmax)
//@function: JsonInBlacklist
//@description: 拉黑jwt
//@param: jwtList model.JwtBlacklist
//@return: err error

func (jwtService *JwtService) JsonInBlacklist(jwtList system.JwtBlacklist) (err error) {
	err = jwtService.db.Create(&jwtList).Error
	if err != nil {
		return
	}
	auth.RevokeToken(jwtList.Jwt)
	return
}

func (jwtService *JwtService) LoadRevokedTokens() {
	var data []string
	err := jwtService.db.Model(&system.JwtBlacklist{}).Select("jwt").Find(&data).Error
	if err != nil {
		infra.Log.Error("加载数据库jwt黑名单失败!", zap.Error(err))
		return
	}
	for _, token := range data {
		auth.RevokeToken(token)
	}
}
