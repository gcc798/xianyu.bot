package response

import "github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model"

type ResourceDetail struct {
	model.Resource
	Sources []model.ResourceSource `json:"sources"`
}
