package system

import (
	"github.com/gcc798/xianyu.bot/internal/model/common"
)

type RevokedToken struct {
	common.Timestamps
	Token string `gorm:"column:jwt;type:text;comment:jwt"`
}

func (RevokedToken) TableName() string { return "jwt_blacklists" }
