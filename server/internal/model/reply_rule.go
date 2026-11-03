package model

const (
	// ReplyRuleExact 表示关键词必须完全匹配。
	ReplyRuleExact = "exact"
	// ReplyRuleFuzzy 表示消息包含关键词即可匹配。
	ReplyRuleFuzzy = "fuzzy"
	// ReplyRuleFallback 表示其他规则未命中时使用。
	ReplyRuleFallback = "fallback"
)

// ReplyRule 定义商品消息触发自动回复的条件。
type ReplyRule struct {
	Timestamps        // Timestamps 保存规则记录的生命周期字段。
	OwnerID    uint   `json:"-" gorm:"not null;index:idx_xy_reply_rules_owner"`                      // OwnerID 是规则所属的系统用户 ID。
	ProductID  uint   `json:"productId" gorm:"not null;uniqueIndex:idx_xy_reply_rule_unique"`        // ProductID 是规则绑定的商品 ID。
	Type       string `json:"type" gorm:"size:16;not null;uniqueIndex:idx_xy_reply_rule_unique"`     // Type 是精确、模糊或兜底匹配类型。
	Keyword    string `json:"keyword" gorm:"size:255;not null;uniqueIndex:idx_xy_reply_rule_unique"` // Keyword 是用于匹配买家消息的关键词。
	Priority   int    `json:"priority" gorm:"not null;default:0"`                                    // Priority 决定多条规则命中时的选择顺序。
	Enabled    bool   `json:"enabled" gorm:"not null"`                                               // Enabled 决定规则是否参与匹配。
}

// TableName 返回自动回复规则表名。
func (ReplyRule) TableName() string { return "xy_reply_rules" }

// ReplyContent 保存规则命中后发送的文本。
type ReplyContent struct {
	Timestamps        // Timestamps 保存回复内容的生命周期字段。
	RuleID     uint   `json:"ruleId" gorm:"not null;uniqueIndex"` // RuleID 是内容所属的回复规则 ID。
	Text       string `json:"text" gorm:"type:text;not null"`     // Text 是发送给买家的回复文本。
}

// TableName 返回自动回复内容表名。
func (ReplyContent) TableName() string { return "xy_reply_contents" }
