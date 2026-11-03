package response

import "time"

// DeliveryRule 展示带商品标题的自动发货规则。
type DeliveryRule struct {
	ID           uint      `json:"ID"`           // ID 是规则 ID。
	CreatedAt    time.Time `json:"CreatedAt"`    // CreatedAt 是规则创建时间。
	UpdatedAt    time.Time `json:"UpdatedAt"`    // UpdatedAt 是规则最后更新时间。
	ProductID    uint      `json:"productId"`    // ProductID 是规则绑定的商品 ID。
	ProductTitle string    `json:"productTitle"` // ProductTitle 是商品标题。
	SKUKey       string    `json:"skuKey"`       // SKUKey 是商品规格稳定标识。
	SKUName      string    `json:"skuName"`      // SKUName 是规格显示名称。
	Mode         string    `json:"mode"`         // Mode 是文本或卡密发货模式。
	Template     string    `json:"template"`     // Template 是发送给买家的消息模板。
	Enabled      bool      `json:"enabled"`      // Enabled 表示规则是否参与自动发货。
}
