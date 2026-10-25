package response

import "time"

type BusinessEvent struct {
	ID              uint      `json:"ID"`
	CreatedAt       time.Time `json:"CreatedAt"`
	AccountID       uint      `json:"accountId"`
	AccountLabel    string    `json:"accountLabel"`
	ProductID       uint      `json:"productId"`
	ProductTitle    string    `json:"productTitle"`
	OrderID         uint      `json:"orderId"`
	ExternalOrderID string    `json:"externalOrderId"`
	Type            string    `json:"type"`
	Status          string    `json:"status"`
	Summary         string    `json:"summary"`
	Error           string    `json:"error"`
}
