package response

import "time"

// Order 展示一条带账号名称的闲鱼订单。
type Order struct {
	// ID 是订单记录 ID。
	ID uint `json:"ID"`
	// CreatedAt 是订单首次入库时间。
	CreatedAt time.Time `json:"CreatedAt"`
	// UpdatedAt 是订单最后同步时间。
	UpdatedAt time.Time `json:"UpdatedAt"`
	// AccountID 是订单所属的闲鱼账号 ID。
	AccountID uint `json:"accountId"`
	// AccountLabel 是账号显示名称。
	AccountLabel string `json:"accountLabel"`
	// ExternalOrderID 是闲鱼侧订单标识。
	ExternalOrderID string `json:"externalOrderId"`
	// ExternalProductID 是闲鱼侧商品标识。
	ExternalProductID string `json:"externalProductId"`
	// ProductTitle 是下单时的商品标题。
	ProductTitle string `json:"productTitle"`
	// BuyerUserID 是闲鱼买家标识。
	BuyerUserID string `json:"buyerUserId"`
	// BuyerUserName 是买家显示名称。
	BuyerUserName string `json:"buyerUserName"`
	// SessionID 是订单关联的会话标识。
	SessionID string `json:"sessionId"`
	// SKUKey 是成交商品规格标识。
	SKUKey string `json:"skuKey"`
	// Status 是归一化后的订单状态。
	Status string `json:"status"`
	// AmountCent 是以分为单位的订单金额。
	AmountCent int64 `json:"amountCent"`
}
