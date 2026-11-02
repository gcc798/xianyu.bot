package request

import (
	"github.com/gcc798/xianyu.bot/internal/model"
)

type LoginLogSearch struct {
	model.LoginLog
	PageInfo
}
