package response

import "time"

// Order 展示一条带账号名称的闲鱼订单。
type Order struct {
	ID                uint      `json:"ID"`                // ID 是订单记录 ID。
	CreatedAt         time.Time `json:"CreatedAt"`         // CreatedAt 是订单首次入库时间。
	UpdatedAt         time.Time `json:"UpdatedAt"`         // UpdatedAt 是订单最后同步时间。
	AccountID         uint      `json:"accountId"`         // AccountID 是订单所属的闲鱼账号 ID。
	AccountLabel      string    `json:"accountLabel"`      // AccountLabel 是账号显示名称。
	ExternalOrderID   string    `json:"externalOrderId"`   // ExternalOrderID 是闲鱼侧订单标识。
	ExternalProductID string    `json:"externalProductId"` // ExternalProductID 是闲鱼侧商品标识。
	ProductTitle      string    `json:"productTitle"`      // ProductTitle 是下单时的商品标题。
	BuyerUserID       string    `json:"buyerUserId"`       // BuyerUserID 是闲鱼买家标识。
	BuyerUserName     string    `json:"buyerUserName"`     // BuyerUserName 是买家显示名称。
	SessionID         string    `json:"sessionId"`         // SessionID 是订单关联的会话标识。
	SKUKey            string    `json:"skuKey"`            // SKUKey 是成交商品规格标识。
	Status            string    `json:"status"`            // Status 是归一化后的订单状态。
	AmountCent        int64     `json:"amountCent"`        // AmountCent 是以分为单位的订单金额。
}
