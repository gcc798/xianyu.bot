package request

import (
	commonRequest "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"gorm.io/gorm"
)

type ResourceSearch struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
	Keyword  string `json:"keyword" form:"keyword"`
	MIMEType string `json:"mimeType" form:"mimeType"`
	State    string `json:"state" form:"state"`
}

type CompleteResourceDelivery struct {
	Size   int64  `json:"size" binding:"required"`
	SHA256 string `json:"sha256" binding:"required"`
}

func (r *ResourceSearch) Paginate() func(*gorm.DB) *gorm.DB {
	return (&commonRequest.PageInfo{Page: r.Page, PageSize: r.PageSize}).Paginate()
}
