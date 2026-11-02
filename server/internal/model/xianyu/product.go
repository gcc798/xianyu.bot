package model

import "github.com/gcc798/go.xianyu.assistant/server/internal/model/common"

type Product struct {
	common.Timestamps
	OwnerID           uint   `json:"-" gorm:"not null;index:idx_xy_products_owner"`
	AccountID         uint   `json:"accountId" gorm:"not null;uniqueIndex:idx_xy_products_account_external"`
	ExternalProductID string `json:"externalProductId" gorm:"size:64;not null;uniqueIndex:idx_xy_products_account_external"`
	Title             string `json:"title" gorm:"size:255;not null"`
	PriceCent         int64  `json:"priceCent" gorm:"not null;default:0"`
	Status            string `json:"status" gorm:"size:32;not null;default:on_sale"`
	ImageURL          string `json:"imageUrl" gorm:"size:2048"`
	PublishSnapshot   string `json:"-" gorm:"type:text"`
	SourceJSON        string `json:"-" gorm:"type:text"`
}

func (Product) TableName() string { return "xy_products" }

type ProductSetting struct {
	common.Timestamps
	ProductID           uint   `json:"productId" gorm:"not null;uniqueIndex"`
	VirtualProduct      bool   `json:"virtualProduct" gorm:"not null;default:false"`
	AutoReplyEnabled    bool   `json:"autoReplyEnabled" gorm:"not null;default:false"`
	AutoDeliveryEnabled bool   `json:"autoDeliveryEnabled" gorm:"not null;default:false"`
	AutoRelistEnabled   bool   `json:"autoRelistEnabled" gorm:"not null;default:false"`
	ReplyDelaySeconds   int    `json:"replyDelaySeconds" gorm:"not null;default:0"`
	RelistDelaySeconds  int    `json:"relistDelaySeconds" gorm:"not null;default:60"`
	RelistDescription   string `json:"relistDescription" gorm:"type:text"`
	RelistImageURLs     string `json:"relistImageUrls" gorm:"type:text"`
	RelistAddress       string `json:"relistAddress" gorm:"size:255"`
}

func (ProductSetting) TableName() string { return "xy_product_settings" }
