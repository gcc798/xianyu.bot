package model

const (
	// DeliveryModeText 表示直接发送文本模板。
	DeliveryModeText = "text"
	// DeliveryModeCard 表示从卡密库存中领取并发送资源。
	DeliveryModeCard = "card"
)

// DeliveryRule 定义商品 SKU 成交后的自动发货方式。
type DeliveryRule struct {
	Timestamps        // Timestamps 保存规则记录的生命周期字段。
	OwnerID    uint   `json:"-" gorm:"not null;index:idx_xy_delivery_rules_owner"`                          // OwnerID 是规则所属的系统用户 ID。
	ProductID  uint   `json:"productId" gorm:"not null;uniqueIndex:idx_xy_delivery_rule_product_sku"`       // ProductID 是规则绑定的商品 ID。
	SKUKey     string `json:"skuKey" gorm:"size:128;not null;uniqueIndex:idx_xy_delivery_rule_product_sku"` // SKUKey 是商品规格的稳定标识。
	SKUName    string `json:"skuName" gorm:"size:255"`                                                      // SKUName 是规格展示名称。
	Mode       string `json:"mode" gorm:"size:16;not null"`                                                 // Mode 是文本或卡密发货模式。
	Template   string `json:"template" gorm:"type:text;not null"`                                           // Template 是发送给买家的消息模板。
	Enabled    bool   `json:"enabled" gorm:"not null"`                                                      // Enabled 决定规则是否参与自动发货。
}

// TableName 返回自动发货规则表名。
func (DeliveryRule) TableName() string { return "xy_delivery_rules" }
