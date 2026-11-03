package model

// Setting 保存一个服务级键值配置。
type Setting struct {
	Key   string `gorm:"primaryKey;size:128"` // Key 是配置项的唯一标识。
	Value string `gorm:"type:text;not null"`  // Value 是配置项的文本值。
}

// TableName 返回服务设置表名。
func (Setting) TableName() string { return "xy_settings" }
