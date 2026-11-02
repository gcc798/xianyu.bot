package request

import commonRequest "github.com/gcc798/xianyu.bot/internal/model/common/request"

type MessageSearch struct {
	commonRequest.Pagination
	AccountID         uint   `json:"accountId" form:"accountId"`
	SessionID         string `json:"sessionId" form:"sessionId"`
	Direction         string `json:"direction" form:"direction"`
	ExternalProductID string `json:"externalProductId" form:"externalProductId"`
	BuyerName         string `json:"buyerName" form:"buyerName"`
	Content           string `json:"content" form:"content"`
	StartTime         int64  `json:"startTime" form:"startTime"`
	EndTime           int64  `json:"endTime" form:"endTime"`
}

type SendMessage struct {
	MessageID uint   `json:"messageId"`
	Text      string `json:"text"`
}

type UpdateTakeover struct {
	MessageID uint `json:"messageId"`
	Active    bool `json:"active"`
}
