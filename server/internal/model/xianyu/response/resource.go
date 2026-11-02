package response

import "github.com/gcc798/go.xianyu.assistant/server/internal/model/xianyu"

type ResourceDetail struct {
	model.Resource
	Sources []model.ResourceSource `json:"sources"`
}
