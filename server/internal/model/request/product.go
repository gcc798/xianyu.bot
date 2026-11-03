package request

// ProductSearch 描述商品列表的分页与筛选条件。
type ProductSearch struct {
	// Pagination 提供页码与每页数量。
	Pagination
	// AccountID 仅返回指定闲鱼账号的商品。
	AccountID uint `json:"accountId" form:"accountId"`
	// TitleOrID 按商品标题或闲鱼商品标识检索。
	TitleOrID string `json:"titleOrId" form:"titleOrId"`
	// Status 仅返回指定状态的商品。
	Status string `json:"status" form:"status"`
}

// ProductSync 描述手动同步商品的目标账号。
type ProductSync struct {
	// AccountID 是待同步的闲鱼账号 ID。
	AccountID uint `json:"accountId"`
}

// UpdateProductSetting 描述商品自动化配置更新。
type UpdateProductSetting struct {
	// ProductID 是目标本地商品 ID。
	ProductID uint `json:"productId"`
	// VirtualProduct 标识商品是否无需物流发货。
	VirtualProduct bool `json:"virtualProduct"`
	// AutoReplyEnabled 决定是否自动回复买家消息。
	AutoReplyEnabled bool `json:"autoReplyEnabled"`
	// AutoDeliveryEnabled 决定成交后是否自动发货。
	AutoDeliveryEnabled bool `json:"autoDeliveryEnabled"`
	// AutoRelistEnabled 决定售出后是否自动重发商品。
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
