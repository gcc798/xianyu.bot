package response

import "time"

// Product 展示商品快照及其自动化设置。
type Product struct {
	ID                  uint      `json:"ID"`                  // ID 是商品记录 ID。
	CreatedAt           time.Time `json:"CreatedAt"`           // CreatedAt 是商品首次入库时间。
	UpdatedAt           time.Time `json:"UpdatedAt"`           // UpdatedAt 是商品最后同步时间。
	AccountID           uint      `json:"accountId"`           // AccountID 是商品所属的闲鱼账号 ID。
	AccountLabel        string    `json:"accountLabel"`        // AccountLabel 是账号显示名称。
	ExternalProductID   string    `json:"externalProductId"`   // ExternalProductID 是闲鱼侧商品标识。
	Title               string    `json:"title"`               // Title 是商品标题。
	PriceCent           int64     `json:"priceCent"`           // PriceCent 是以分为单位的商品价格。
	Status              string    `json:"status"`              // Status 是归一化后的商品状态。
	ImageURL            string    `json:"imageUrl"`            // ImageURL 是商品主图地址。
	ImageURLs           []string  `json:"imageUrls"`           // ImageURLs 是商品全部图片地址。
	SourceJSON          string    `json:"-"`                   // SourceJSON 是闲鱼原始商品数据，不向客户端返回。
	VirtualProduct      bool      `json:"virtualProduct"`      // VirtualProduct 标识商品是否无需物流发货。
	AutoReplyEnabled    bool      `json:"autoReplyEnabled"`    // AutoReplyEnabled 表示是否启用自动回复。
	AutoDeliveryEnabled bool      `json:"autoDeliveryEnabled"` // AutoDeliveryEnabled 表示是否启用自动发货。
	AutoRelistEnabled   bool      `json:"autoRelistEnabled"`   // AutoRelistEnabled 表示是否启用自动重发。
	ReplyDelaySeconds   int       `json:"replyDelaySeconds"`   // ReplyDelaySeconds 是自动回复前的等待秒数。
	RelistDelaySeconds  int       `json:"relistDelaySeconds"`  // RelistDelaySeconds 是创建重发任务前的等待秒数。
	RelistDescription   string    `json:"relistDescription"`   // RelistDescription 是重发时覆盖的商品描述。
	RelistImageURLs     string    `json:"relistImageUrls"`     // RelistImageURLs 是重发图片地址集合的序列化值。
	RelistAddress       string    `json:"relistAddress"`       // RelistAddress 是重发商品的所在地。
}
