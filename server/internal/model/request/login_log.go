package request

import (
	"github.com/gcc798/xianyu.bot/internal/model"
)

// LoginLogSearch 描述登录审计分页与字段筛选条件。
type LoginLogSearch struct {
	model.LoginLog // LoginLog 提供登录日志字段筛选值。
	PageInfo       // PageInfo 提供分页和关键词条件。
}
