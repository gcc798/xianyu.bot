package response

import "github.com/gcc798/xianyu.bot/internal/model/xianyu"

type ResourceDetail struct {
	model.Resource
	Sources []model.ResourceSource `json:"sources"`
}
