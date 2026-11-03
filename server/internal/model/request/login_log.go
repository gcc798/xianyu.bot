package request

import (
	"github.com/gcc798/xianyu.bot/internal/model"
)

// LoginLogSearch 描述登录审计分页与字段筛选条件。
type LoginLogSearch struct {
	// LoginLog 提供登录日志字段筛选值。
	model.LoginLog
	// PageInfo 提供分页和关键词条件。
	PageInfo
}
