package request

import (
	"github.com/gcc798/xianyu.bot/internal/model"
)

type DictionaryDetailSearch struct {
	model.DictionaryDetail
	PageInfo
}
