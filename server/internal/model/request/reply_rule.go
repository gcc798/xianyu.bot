package request

// ReplyRuleSearch 描述自动回复规则列表的筛选条件。
type ReplyRuleSearch struct {
	PageInfo         // PageInfo 提供分页和关键词条件。
	ProductID uint   `json:"productId" form:"productId"` // ProductID 仅返回指定商品的规则。
	Type      string `json:"type" form:"type"`           // Type 仅返回指定匹配类型的规则。
}

// SaveReplyRule 描述创建或更新自动回复规则的输入。
type SaveReplyRule struct {
	ID        uint   `json:"ID"`        // ID 是更新时的规则 ID，新建时为零。
	ProductID uint   `json:"productId"` // ProductID 是规则绑定的商品 ID。
	Type      string `json:"type"`      // Type 是精确、模糊或兜底匹配类型。
	Keyword   string `json:"keyword"`   // Keyword 是用于匹配买家消息的关键词。
	Priority  int    `json:"priority"`  // Priority 决定多条规则命中时的选择顺序。
	Enabled   bool   `json:"enabled"`   // Enabled 决定规则是否参与匹配。
	Content   string `json:"content"`   // Content 是规则命中后发送的文本。
}

// ReplyRuleID 标识一条自动回复规则。
type ReplyRuleID struct {
	ID uint `json:"ID" form:"ID"` // ID 是目标规则 ID。
}
