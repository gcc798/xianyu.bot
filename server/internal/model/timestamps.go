package model

import (
	"time"

	"gorm.io/gorm"
)

// Timestamps 为持久化模型提供主键、审计时间和软删除字段。
type Timestamps struct {
	// ID 是数据库自增主键。
	ID uint `gorm:"primarykey" json:"ID"`
	// CreatedAt 是记录创建时间。
	CreatedAt time.Time
	// UpdatedAt 是记录最后更新时间。
	UpdatedAt time.Time
	// DeletedAt 标记软删除时间，不向客户端返回。
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
