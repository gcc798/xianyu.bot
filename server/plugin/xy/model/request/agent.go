package request

import (
	commonRequest "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"gorm.io/gorm"
)

type CreateAgentSession struct {
	Title string `json:"title" binding:"max=200"`
}

type AgentSessionSearch struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"pageSize" form:"pageSize"`
}

func (r *AgentSessionSearch) Paginate() func(*gorm.DB) *gorm.DB {
	return (&commonRequest.PageInfo{Page: r.Page, PageSize: r.PageSize}).Paginate()
}
