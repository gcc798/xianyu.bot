package response

import "github.com/gcc798/xianyu.bot/internal/model"

// ResourceDetail 展示本地资源及其全部外部来源。
type ResourceDetail struct {
	model.Resource                        // Resource 是本地资源基本信息。
	Sources        []model.ResourceSource `json:"sources"` // Sources 是可审计的外部来源记录。
}
