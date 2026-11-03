package request

import (
	"github.com/gcc798/xianyu.bot/internal/model"
)

// DictionaryDetailSearch 描述字典条目分页与字段筛选条件。
type DictionaryDetailSearch struct {
	model.DictionaryDetail // DictionaryDetail 提供条目字段筛选值。
	PageInfo               // PageInfo 提供分页和关键词条件。
}
