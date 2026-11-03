package response

import "time"

// ReplyRule 展示带商品标题和回复内容的自动回复规则。
type ReplyRule struct {
	ID           uint      `json:"ID"`           // ID 是规则 ID。
	CreatedAt    time.Time `json:"CreatedAt"`    // CreatedAt 是规则创建时间。
	UpdatedAt    time.Time `json:"UpdatedAt"`    // UpdatedAt 是规则最后更新时间。
	ProductID    uint      `json:"productId"`    // ProductID 是规则绑定的商品 ID。
	ProductTitle string    `json:"productTitle"` // ProductTitle 是商品标题。
	Type         string    `json:"type"`         // Type 是精确、模糊或兜底匹配类型。
	Keyword      string    `json:"keyword"`      // Keyword 是用于匹配买家消息的关键词。
	Priority     int       `json:"priority"`     // Priority 决定多条规则命中时的选择顺序。
	Enabled      bool      `json:"enabled"`      // Enabled 表示规则是否参与匹配。
	Content      string    `json:"content"`      // Content 是规则命中后发送的文本。
}
