package response

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model"

type ResourceDetail struct {
	model.Resource
	Sources []model.ResourceSource `json:"sources"`
}
