package response

import "time"

type Order struct {
	ID                uint      `json:"ID"`
	CreatedAt         time.Time `json:"CreatedAt"`
	UpdatedAt         time.Time `json:"UpdatedAt"`
	AccountID         uint      `json:"accountId"`
	AccountLabel      string    `json:"accountLabel"`
	ExternalOrderID   string    `json:"externalOrderId"`
	ExternalProductID string    `json:"externalProductId"`
	ProductTitle      string    `json:"productTitle"`
	BuyerUserID       string    `json:"buyerUserId"`
	BuyerUserName     string    `json:"buyerUserName"`
	SessionID         string    `json:"sessionId"`
	SKUKey            string    `json:"skuKey"`
	Status            string    `json:"status"`
	AmountCent        int64     `json:"amountCent"`
}
