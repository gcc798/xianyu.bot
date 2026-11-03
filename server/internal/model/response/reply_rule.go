package response

import "time"

// ReplyRule 展示带商品标题和回复内容的自动回复规则。
type ReplyRule struct {
	// ID 是规则 ID。
	ID uint `json:"ID"`
	// CreatedAt 是规则创建时间。
	CreatedAt time.Time `json:"CreatedAt"`
	// UpdatedAt 是规则最后更新时间。
	UpdatedAt time.Time `json:"UpdatedAt"`
	// ProductID 是规则绑定的商品 ID。
	ProductID uint `json:"productId"`
	// ProductTitle 是商品标题。
	ProductTitle string `json:"productTitle"`
	// Type 是精确、模糊或兜底匹配类型。
	Type string `json:"type"`
	// Keyword 是用于匹配买家消息的关键词。
	Keyword string `json:"keyword"`
	// Priority 决定多条规则命中时的选择顺序。
	Priority int `json:"priority"`
	// Enabled 表示规则是否参与匹配。
	Enabled bool `json:"enabled"`
	// Content 是规则命中后发送的文本。
	Content string `json:"content"`
}
