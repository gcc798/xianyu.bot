package model

import (
	"time"

	"gorm.io/gorm"
)

// Timestamps 为持久化模型提供主键、审计时间和软删除字段。
type Timestamps struct {
	ID        uint           `gorm:"primarykey" json:"ID"` // ID 是数据库自增主键。
	CreatedAt time.Time      // CreatedAt 是记录创建时间。
	UpdatedAt time.Time      // UpdatedAt 是记录最后更新时间。
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // DeletedAt 标记软删除时间，不向客户端返回。
}
