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

type UpdateResourceDelivery struct {
	Stage           string `json:"stage" binding:"required,max=64"`
	ProgressPercent int    `json:"progressPercent" binding:"min=0,max=100"`
	ProgressDone    int64  `json:"progressDone" binding:"min=0"`
	ProgressTotal   int64  `json:"progressTotal" binding:"min=0"`
	ProgressUnit    string `json:"progressUnit" binding:"max=32"`
}

type FailResourceDelivery struct {
	ErrorCode    string `json:"errorCode" binding:"required,max=64"`
	ErrorMessage string `json:"errorMessage" binding:"required,max=2000"`
}

func (r *ResourceSearch) Paginate() func(*gorm.DB) *gorm.DB {
	return (&commonRequest.PageInfo{Page: r.Page, PageSize: r.PageSize}).Paginate()
}
