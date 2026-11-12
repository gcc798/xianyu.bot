package request

// ProductSearch 描述商品列表的分页与筛选条件。
type ProductSearch struct {
	Pagination        // Pagination 提供页码与每页数量。
	AccountID  uint   `json:"accountId" form:"accountId"` // AccountID 仅返回指定闲鱼账号的商品。
	TitleOrID  string `json:"titleOrId" form:"titleOrId"` // TitleOrID 按商品标题或闲鱼商品标识检索。
	Status     string `json:"status" form:"status"`       // Status 仅返回指定状态的商品。
}

// ProductSync 描述手动同步商品的目标账号。
type ProductSync struct {
	AccountID uint `json:"accountId"` // AccountID 是待同步的闲鱼账号 ID。
}

// UpdateProductSetting 描述商品自动化配置更新。
type UpdateProductSetting struct {
	ProductID           uint   `json:"productId"`           // ProductID 是目标本地商品 ID。
	VirtualProduct      bool   `json:"virtualProduct"`      // VirtualProduct 标识商品是否无需物流发货。
	AutoReplyEnabled    bool   `json:"autoReplyEnabled"`    // AutoReplyEnabled 决定是否自动回复买家消息。
	AutoDeliveryEnabled bool   `json:"autoDeliveryEnabled"` // AutoDeliveryEnabled 决定成交后是否自动发货。
	AutoRelistEnabled   bool   `json:"autoRelistEnabled"`   // AutoRelistEnabled 决定售出后是否自动重发商品。
	ReplyDelaySeconds   int    `json:"replyDelaySeconds"`   // ReplyDelaySeconds 是自动回复前的等待秒数。
	RelistDelaySeconds  int    `json:"relistDelaySeconds"`  // RelistDelaySeconds 是创建重发任务前的等待秒数。
	RelistDescription   string `json:"relistDescription"`   // RelistDescription 是重发时覆盖的商品描述。
	RelistImageURLs     string `json:"relistImageUrls"`     // RelistImageURLs 是重发图片地址集合的序列化值。
	RelistAddress       string `json:"relistAddress"`       // RelistAddress 是重发商品的所在地。
}

// PatchProductSetting 描述只更新明确提供字段的商品自动化配置。
type PatchProductSetting struct {
	ProductID           uint    `json:"productId"`                     // ProductID 是目标本地商品 ID。
	VirtualProduct      *bool   `json:"virtualProduct,omitempty"`      // VirtualProduct 更新虚拟商品标识。
	AutoReplyEnabled    *bool   `json:"autoReplyEnabled,omitempty"`    // AutoReplyEnabled 更新自动回复开关。
	AutoDeliveryEnabled *bool   `json:"autoDeliveryEnabled,omitempty"` // AutoDeliveryEnabled 更新自动发货开关。
	AutoRelistEnabled   *bool   `json:"autoRelistEnabled,omitempty"`   // AutoRelistEnabled 更新自动续售开关。
	ReplyDelaySeconds   *int    `json:"replyDelaySeconds,omitempty"`   // ReplyDelaySeconds 更新回复延迟。
	RelistDelaySeconds  *int    `json:"relistDelaySeconds,omitempty"`  // RelistDelaySeconds 更新续售延迟。
	RelistDescription   *string `json:"relistDescription,omitempty"`   // RelistDescription 更新续售描述。
	RelistImageURLs     *string `json:"relistImageUrls,omitempty"`     // RelistImageURLs 更新续售图片。
	RelistAddress       *string `json:"relistAddress,omitempty"`       // RelistAddress 更新续售所在地。
}
