package response

import "time"

type DeliveryRule struct {
	ID           uint      `json:"ID"`
	CreatedAt    time.Time `json:"CreatedAt"`
	UpdatedAt    time.Time `json:"UpdatedAt"`
	ProductID    uint      `json:"productId"`
	ProductTitle string    `json:"productTitle"`
	SKUKey       string    `json:"skuKey"`
	SKUName      string    `json:"skuName"`
	Mode         string    `json:"mode"`
	Template     string    `json:"template"`
	Enabled      bool      `json:"enabled"`
}
