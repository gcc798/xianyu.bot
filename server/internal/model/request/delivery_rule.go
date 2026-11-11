package request

// DeliveryRuleSearch 描述自动发货规则列表的筛选条件。
type DeliveryRuleSearch struct {
	PageInfo           // PageInfo 提供分页和关键词条件。
	ProductID   uint   `json:"productId" form:"productId"`     // ProductID 仅返回指定商品的规则。
	Mode        string `json:"mode" form:"mode"`               // Mode 仅返回指定发货模式的规则。
	EnabledOnly bool   `json:"enabledOnly" form:"enabledOnly"` // EnabledOnly 仅返回启用规则。
}

// SaveDeliveryRule 描述创建或更新自动发货规则的输入。
type SaveDeliveryRule struct {
	ID        uint   `json:"ID"`        // ID 是更新时的规则 ID，新建时为零。
	ProductID uint   `json:"productId"` // ProductID 是规则绑定的商品 ID。
	SKUKey    string `json:"skuKey"`    // SKUKey 是商品规格稳定标识。
	SKUName   string `json:"skuName"`   // SKUName 是规格显示名称。
	Mode      string `json:"mode"`      // Mode 是文本或卡密发货模式。
	Template  string `json:"template"`  // Template 是发送给买家的消息模板。
	Enabled   bool   `json:"enabled"`   // Enabled 决定规则是否参与自动发货。
}

// DeliveryRuleID 标识一条自动发货规则。
type DeliveryRuleID struct {
	ID uint `json:"ID" form:"ID"` // ID 是目标规则 ID。
}
