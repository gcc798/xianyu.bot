package request

import (
	"github.com/gcc798/xianyu.bot/internal/model/common/request"
	"github.com/gcc798/xianyu.bot/internal/model/system"
)

type SysDictionaryDetailSearch struct {
	system.SysDictionaryDetail
	request.PageInfo
}
