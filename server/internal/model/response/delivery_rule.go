package response

import "time"

// DeliveryRule 展示带商品标题的自动发货规则。
type DeliveryRule struct {
	// ID 是规则 ID。
	ID uint `json:"ID"`
	// CreatedAt 是规则创建时间。
	CreatedAt time.Time `json:"CreatedAt"`
	// UpdatedAt 是规则最后更新时间。
	UpdatedAt time.Time `json:"UpdatedAt"`
	// ProductID 是规则绑定的商品 ID。
	ProductID uint `json:"productId"`
	// ProductTitle 是商品标题。
	ProductTitle string `json:"productTitle"`
	// SKUKey 是商品规格稳定标识。
	SKUKey string `json:"skuKey"`
	// SKUName 是规格显示名称。
	SKUName string `json:"skuName"`
	// Mode 是文本或卡密发货模式。
	Mode string `json:"mode"`
	// Template 是发送给买家的消息模板。
	Template string `json:"template"`
	// Enabled 表示规则是否参与自动发货。
	Enabled bool `json:"enabled"`
}
