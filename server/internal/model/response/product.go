package response

import "time"

// Product 展示商品快照及其自动化设置。
type Product struct {
	// ID 是商品记录 ID。
	ID uint `json:"ID"`
	// CreatedAt 是商品首次入库时间。
	CreatedAt time.Time `json:"CreatedAt"`
	// UpdatedAt 是商品最后同步时间。
	UpdatedAt time.Time `json:"UpdatedAt"`
	// AccountID 是商品所属的闲鱼账号 ID。
	AccountID uint `json:"accountId"`
	// AccountLabel 是账号显示名称。
	AccountLabel string `json:"accountLabel"`
	// ExternalProductID 是闲鱼侧商品标识。
	ExternalProductID string `json:"externalProductId"`
	// Title 是商品标题。
	Title string `json:"title"`
	// PriceCent 是以分为单位的商品价格。
	PriceCent int64 `json:"priceCent"`
	// Status 是归一化后的商品状态。
	Status string `json:"status"`
	// ImageURL 是商品主图地址。
	ImageURL string `json:"imageUrl"`
	// ImageURLs 是商品全部图片地址。
	ImageURLs []string `json:"imageUrls"`
	// SourceJSON 是闲鱼原始商品数据，不向客户端返回。
	SourceJSON string `json:"-"`
	// VirtualProduct 标识商品是否无需物流发货。
	VirtualProduct bool `json:"virtualProduct"`
	// AutoReplyEnabled 表示是否启用自动回复。
	AutoReplyEnabled bool `json:"autoReplyEnabled"`
	// AutoDeliveryEnabled 表示是否启用自动发货。
	AutoDeliveryEnabled bool `json:"autoDeliveryEnabled"`
	// AutoRelistEnabled 表示是否启用自动重发。
	AutoRelistEnabled bool `json:"autoRelistEnabled"`
	// ReplyDelaySeconds 是自动回复前的等待秒数。
	ReplyDelaySeconds int `json:"replyDelaySeconds"`
	// RelistDelaySeconds 是创建重发任务前的等待秒数。
	RelistDelaySeconds int `json:"relistDelaySeconds"`
	// RelistDescription 是重发时覆盖的商品描述。
	RelistDescription string `json:"relistDescription"`
	// RelistImageURLs 是重发图片地址集合的序列化值。
	RelistImageURLs string `json:"relistImageUrls"`
	// RelistAddress 是重发商品的所在地。
	RelistAddress string `json:"relistAddress"`
}
