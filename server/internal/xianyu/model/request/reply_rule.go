package request

import commonRequest "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type ReplyRuleSearch struct {
	commonRequest.PageInfo
	ProductID uint   `json:"productId" form:"productId"`
	Type      string `json:"type" form:"type"`
}

type SaveReplyRule struct {
	ID        uint   `json:"ID"`
	ProductID uint   `json:"productId"`
	Type      string `json:"type"`
	Keyword   string `json:"keyword"`
	Priority  int    `json:"priority"`
	Enabled   bool   `json:"enabled"`
	Content   string `json:"content"`
}

type ReplyRuleID struct {
	ID uint `json:"ID" form:"ID"`
}
