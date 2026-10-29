package model

import "github.com/flipped-aurora/gin-vue-admin/server/global"

const (
	DeliveryModeText = "text"
	DeliveryModeCard = "card"
)

type DeliveryRule struct {
	global.GVA_MODEL
	OwnerID   uint   `json:"-" gorm:"not null;index:idx_xy_delivery_rules_owner"`
	ProductID uint   `json:"productId" gorm:"not null;uniqueIndex:idx_xy_delivery_rule_product_sku"`
	SKUKey    string `json:"skuKey" gorm:"size:128;not null;uniqueIndex:idx_xy_delivery_rule_product_sku"`
	SKUName   string `json:"skuName" gorm:"size:255"`
	Mode      string `json:"mode" gorm:"size:16;not null"`
	Template  string `json:"template" gorm:"type:text;not null"`
	Enabled   bool   `json:"enabled" gorm:"not null"`
}

func (DeliveryRule) TableName() string { return "xy_delivery_rules" }
