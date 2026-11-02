package request

import (
	commonRequest "github.com/gcc798/go.xianyu.assistant/server/internal/model/common/request"
	"gorm.io/gorm"
)

type MessageSearch struct {
	Page              int    `json:"page" form:"page"`
	PageSize          int    `json:"pageSize" form:"pageSize"`
	AccountID         uint   `json:"accountId" form:"accountId"`
	SessionID         string `json:"sessionId" form:"sessionId"`
	Direction         string `json:"direction" form:"direction"`
	ExternalProductID string `json:"externalProductId" form:"externalProductId"`
	BuyerName         string `json:"buyerName" form:"buyerName"`
	Content           string `json:"content" form:"content"`
	StartTime         int64  `json:"startTime" form:"startTime"`
	EndTime           int64  `json:"endTime" form:"endTime"`
}

func (r *MessageSearch) Paginate() func(*gorm.DB) *gorm.DB {
	return (&commonRequest.PageInfo{Page: r.Page, PageSize: r.PageSize}).Paginate()
}

type SendMessage struct {
	MessageID uint   `json:"messageId"`
	Text      string `json:"text"`
}

type UpdateTakeover struct {
	MessageID uint `json:"messageId"`
	Active    bool `json:"active"`
}
