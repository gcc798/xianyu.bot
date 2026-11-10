package model

import "time"

// TelegramAPIConfig 保存 Telegram 应用凭据与限流状态。
type TelegramAPIConfig struct {
	Timestamps
	Name                 string     `json:"name" gorm:"size:100;not null;uniqueIndex"`      // Name 是配置名称。
	APIID                int64      `json:"apiId" gorm:"not null;uniqueIndex"`              // APIID 是 Telegram 应用 ID。
	APIHash              string     `json:"-" gorm:"column:api_hash;type:text;not null"`    // APIHash 是加密后的 Telegram 应用凭据。
	FloodWaitUntil       *time.Time `json:"floodWaitUntil"`                                 // FloodWaitUntil 是当前限流结束时间。
	LastFloodWaitSeconds int64      `json:"lastFloodWaitSeconds" gorm:"not null;default:0"` // LastFloodWaitSeconds 是最近一次限流秒数。
	HasAPIHash           bool       `json:"hasApiHash" gorm:"-"`                            // HasAPIHash 表示凭据已配置。
}

// TableName 返回 Telegram API 配置表名。
func (TelegramAPIConfig) TableName() string { return "s_telegram_api_configs" }
