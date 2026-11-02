package request

import (
	"github.com/gcc798/go.xianyu.assistant/server/internal/model/common/request"
	"github.com/gcc798/go.xianyu.assistant/server/internal/model/system"
)

type SysDictionaryDetailSearch struct {
	system.SysDictionaryDetail
	request.PageInfo
}
