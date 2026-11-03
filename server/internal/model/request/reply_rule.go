package request

// ReplyRuleSearch 描述自动回复规则列表的筛选条件。
type ReplyRuleSearch struct {
	// PageInfo 提供分页和关键词条件。
	PageInfo
	// ProductID 仅返回指定商品的规则。
	ProductID uint `json:"productId" form:"productId"`
	// Type 仅返回指定匹配类型的规则。
	Type string `json:"type" form:"type"`
}

// SaveReplyRule 描述创建或更新自动回复规则的输入。
type SaveReplyRule struct {
	// ID 是更新时的规则 ID，新建时为零。
	ID uint `json:"ID"`
	// ProductID 是规则绑定的商品 ID。
	ProductID uint `json:"productId"`
	// Type 是精确、模糊或兜底匹配类型。
	Type string `json:"type"`
	// Keyword 是用于匹配买家消息的关键词。
	Keyword string `json:"keyword"`
	// Priority 决定多条规则命中时的选择顺序。
	Priority int `json:"priority"`
	// Enabled 决定规则是否参与匹配。
	Enabled bool `json:"enabled"`
	// Content 是规则命中后发送的文本。
	Content string `json:"content"`
}

// ReplyRuleID 标识一条自动回复规则。
type ReplyRuleID struct {
	// ID 是目标规则 ID。
	ID uint `json:"ID" form:"ID"`
}
