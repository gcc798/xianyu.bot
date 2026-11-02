package request

import commonRequest "github.com/gcc798/xianyu.bot/internal/model/common/request"

type ProductSearch struct {
	commonRequest.Pagination
	AccountID uint   `json:"accountId" form:"accountId"`
	TitleOrID string `json:"titleOrId" form:"titleOrId"`
	Status    string `json:"status" form:"status"`
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
