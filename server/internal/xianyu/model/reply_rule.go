package model

import "github.com/flipped-aurora/gin-vue-admin/server/global"

const (
	ReplyRuleExact    = "exact"
	ReplyRuleFuzzy    = "fuzzy"
	ReplyRuleFallback = "fallback"
)

type ReplyRule struct {
	global.GVA_MODEL
	OwnerID   uint   `json:"-" gorm:"not null;index:idx_xy_reply_rules_owner"`
	ProductID uint   `json:"productId" gorm:"not null;uniqueIndex:idx_xy_reply_rule_unique"`
	Type      string `json:"type" gorm:"size:16;not null;uniqueIndex:idx_xy_reply_rule_unique"`
	Keyword   string `json:"keyword" gorm:"size:255;not null;uniqueIndex:idx_xy_reply_rule_unique"`
	Priority  int    `json:"priority" gorm:"not null;default:0"`
	Enabled   bool   `json:"enabled" gorm:"not null"`
}

func (ReplyRule) TableName() string { return "xy_reply_rules" }

type ReplyContent struct {
	global.GVA_MODEL
	RuleID uint   `json:"ruleId" gorm:"not null;uniqueIndex"`
	Text   string `json:"text" gorm:"type:text;not null"`
}

func (ReplyContent) TableName() string { return "xy_reply_contents" }
