package system

import (
	"github.com/gcc798/go.xianyu.assistant/server/internal/model/common"
)

type JwtBlacklist struct {
	common.Timestamps
	Jwt string `gorm:"type:text;comment:jwt"`
}
