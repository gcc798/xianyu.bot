package request

import (
	commonRequest "github.com/gcc798/go.xianyu.assistant/server/internal/model/common/request"
	"gorm.io/gorm"
)

type ProductSearch struct {
	Page      int    `json:"page" form:"page"`
	PageSize  int    `json:"pageSize" form:"pageSize"`
	AccountID uint   `json:"accountId" form:"accountId"`
	TitleOrID string `json:"titleOrId" form:"titleOrId"`
	Status    string `json:"status" form:"status"`
}

func (r *ProductSearch) Paginate() func(*gorm.DB) *gorm.DB {
	return (&commonRequest.PageInfo{Page: r.Page, PageSize: r.PageSize}).Paginate()
}

type ProductSync struct {
	AccountID uint `json:"accountId"`
}

type UpdateProductSetting struct {
	ProductID           uint   `json:"productId"`
	VirtualProduct      bool   `json:"virtualProduct"`
	AutoReplyEnabled    bool   `json:"autoReplyEnabled"`
	AutoDeliveryEnabled bool   `json:"autoDeliveryEnabled"`
	AutoRelistEnabled   bool   `json:"autoRelistEnabled"`
	ReplyDelaySeconds   int    `json:"replyDelaySeconds"`
	RelistDelaySeconds  int    `json:"relistDelaySeconds"`
	RelistDescription   string `json:"relistDescription"`
	RelistImageURLs     string `json:"relistImageUrls"`
	RelistAddress       string `json:"relistAddress"`
}
