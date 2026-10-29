package response

import "time"

type Message struct {
	ID                uint      `json:"ID"`
	CreatedAt         time.Time `json:"CreatedAt"`
	AccountID         uint      `json:"accountId"`
	AccountLabel      string    `json:"accountLabel"`
	PNMID             string    `json:"pnmId"`
	SessionID         string    `json:"sessionId"`
	ContentType       int       `json:"contentType"`
	Content           string    `json:"content"`
	SenderUserName    string    `json:"senderUserName"`
	SenderUserID      string    `json:"senderUserId"`
	ExternalProductID string    `json:"externalProductId"`
	ExternalOrderID   string    `json:"externalOrderId"`
	MessageTime       int64     `json:"messageTime"`
	Direction         string    `json:"direction"`
	Source            string    `json:"source"`
	TakeoverActive    bool      `json:"takeoverActive"`
}
