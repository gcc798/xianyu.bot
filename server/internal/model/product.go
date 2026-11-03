package model

// Product 保存从闲鱼同步的商品快照。
type Product struct {
	// Timestamps 保存商品记录的生命周期字段。
	Timestamps
	// OwnerID 是商品所属的系统用户 ID。
	OwnerID uint `json:"-" gorm:"not null;index:idx_xy_products_owner"`
	// AccountID 是商品所属的闲鱼账号 ID。
	AccountID uint `json:"accountId" gorm:"not null;uniqueIndex:idx_xy_products_account_external"`
	// ExternalProductID 是闲鱼侧商品标识。
	ExternalProductID string `json:"externalProductId" gorm:"size:64;not null;uniqueIndex:idx_xy_products_account_external"`
	// Title 是商品标题。
	Title string `json:"title" gorm:"size:255;not null"`
	// PriceCent 是以分为单位的商品价格。
	PriceCent int64 `json:"priceCent" gorm:"not null;default:0"`
	// Status 是归一化后的商品状态。
	Status string `json:"status" gorm:"size:32;not null;default:on_sale"`
	// ImageURL 是商品主图地址。
	ImageURL string `json:"imageUrl" gorm:"size:2048"`
	// PublishSnapshot 保存重新发布所需的商品快照。
	PublishSnapshot string `json:"-" gorm:"type:text"`
	// SourceJSON 保存闲鱼返回的原始商品数据。
	SourceJSON string `json:"-" gorm:"type:text"`
}

// TableName 返回闲鱼商品表名。
func (Product) TableName() string { return "xy_products" }

// ProductSetting 保存单个商品的自动化开关与参数。
type ProductSetting struct {
	// Timestamps 保存配置记录的生命周期字段。
	Timestamps
	// ProductID 是配置对应的本地商品 ID。
	ProductID uint `json:"productId" gorm:"not null;uniqueIndex"`
	// VirtualProduct 标识商品是否无需物流发货。
	VirtualProduct bool `json:"virtualProduct" gorm:"not null;default:false"`
	// AutoReplyEnabled 决定是否自动回复买家消息。
	AutoReplyEnabled bool `json:"autoReplyEnabled" gorm:"not null;default:false"`
	// AutoDeliveryEnabled 决定成交后是否自动发货。
	AutoDeliveryEnabled bool `json:"autoDeliveryEnabled" gorm:"not null;default:false"`
	// AutoRelistEnabled 决定售出后是否自动重发商品。
	AutoRelistEnabled bool `json:"autoRelistEnabled" gorm:"not null;default:false"`
	// ReplyDelaySeconds 是自动回复前的等待秒数。
	ReplyDelaySeconds int `json:"replyDelaySeconds" gorm:"not null;default:0"`
	// RelistDelaySeconds 是创建重发任务前的等待秒数。
	RelistDelaySeconds int `json:"relistDelaySeconds" gorm:"not null;default:60"`
	// RelistDescription 是重发时覆盖的商品描述。
	RelistDescription string `json:"relistDescription" gorm:"type:text"`
	// RelistImageURLs 是重发图片地址集合的序列化值。
	RelistImageURLs string `json:"relistImageUrls" gorm:"type:text"`
	// RelistAddress 是重发商品的所在地。
	RelistAddress string `json:"relistAddress" gorm:"size:255"`
}

// TableName 返回商品自动化设置表名。
func (ProductSetting) TableName() string { return "xy_product_settings" }
