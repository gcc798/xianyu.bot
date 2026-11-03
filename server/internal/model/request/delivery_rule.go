package request

// DeliveryRuleSearch 描述自动发货规则列表的筛选条件。
type DeliveryRuleSearch struct {
	// PageInfo 提供分页和关键词条件。
	PageInfo
	// ProductID 仅返回指定商品的规则。
	ProductID uint `json:"productId" form:"productId"`
	// Mode 仅返回指定发货模式的规则。
	Mode string `json:"mode" form:"mode"`
}

// SaveDeliveryRule 描述创建或更新自动发货规则的输入。
type SaveDeliveryRule struct {
	// ID 是更新时的规则 ID，新建时为零。
	ID uint `json:"ID"`
	// ProductID 是规则绑定的商品 ID。
	ProductID uint `json:"productId"`
	// SKUKey 是商品规格稳定标识。
	SKUKey string `json:"skuKey"`
	// SKUName 是规格显示名称。
	SKUName string `json:"skuName"`
	// Mode 是文本或卡密发货模式。
	Mode string `json:"mode"`
	// Template 是发送给买家的消息模板。
	Template string `json:"template"`
	// Enabled 决定规则是否参与自动发货。
	Enabled bool `json:"enabled"`
}

// DeliveryRuleID 标识一条自动发货规则。
type DeliveryRuleID struct {
	// ID 是目标规则 ID。
	ID uint `json:"ID" form:"ID"`
}
