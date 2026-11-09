package request

import (
	"gorm.io/gorm"
)

// Pagination 描述统一分页参数。
type Pagination struct {
	Page     int `json:"page" form:"page"`         // Page 是从 1 开始的页码。
	PageSize int `json:"pageSize" form:"pageSize"` // PageSize 是每页数量，最大为 100。
}

// Paginate 规范化分页参数并返回对应的 GORM 查询作用域。
func (r *Pagination) Paginate() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if r.Page <= 0 {
			r.Page = 1
		}
		switch {
		case r.PageSize > 100:
			r.PageSize = 100
		case r.PageSize <= 0:
			r.PageSize = 10
		}
		offset := (r.Page - 1) * r.PageSize
		return db.Offset(offset).Limit(r.PageSize)
	}
}

// PageInfo 在统一分页参数上增加关键词检索。
type PageInfo struct {
	Pagination        // Pagination 提供页码与每页数量。
	Keyword    string `json:"keyword" form:"keyword"` // Keyword 是列表模糊检索关键词。
}

// GetById 标识一个使用整数主键的目标记录。
type GetById struct {
	ID int `json:"id" form:"id"` // ID 是目标记录主键。
}
