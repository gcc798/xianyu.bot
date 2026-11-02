package model

type Setting struct {
	Key   string `gorm:"primaryKey;size:128"`
	Value string `gorm:"type:text;not null"`
}

func (Setting) TableName() string { return "xy_settings" }
