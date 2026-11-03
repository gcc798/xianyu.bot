package model

import (
	"time"
)

// Account 表示归属某个系统用户的闲鱼账号。
type Account struct {
	// Timestamps 保存账号记录的生命周期字段。
	Timestamps
	// OwnerID 是账号所属的系统用户 ID。
	OwnerID uint `json:"-" gorm:"not null;index:idx_xy_accounts_owner"`
	// ExternalUserID 是闲鱼侧用户标识。
	ExternalUserID string `json:"externalUserId" gorm:"size:64;index"`
	// DeviceID 是登录设备标识，不向客户端返回。
	DeviceID string `json:"-" gorm:"size:128"`
	// Label 是用户设置的账号显示名称。
	Label string `json:"label" gorm:"size:80;not null"`
	// Note 是账号备注。
	Note string `json:"note" gorm:"size:500"`
	// Status 是账号连接或配置状态。
	Status string `json:"status" gorm:"size:32;not null;default:unconfigured"`
	// Enabled 决定账号是否参与自动化任务。
	Enabled bool `json:"enabled" gorm:"not null;default:true"`
	// HasCredential 表示账号是否已有凭据，仅用于查询投影。
	HasCredential bool `json:"hasCredential" gorm:"->;-:migration"`
}

// TableName 返回闲鱼账号表名。
func (Account) TableName() string { return "xy_accounts" }

// Credential 保存闲鱼账号的加密认证材料。
type Credential struct {
	// Timestamps 保存凭据记录的生命周期字段。
	Timestamps
	// AccountID 是凭据对应的闲鱼账号 ID。
	AccountID uint `json:"-" gorm:"not null;uniqueIndex"`
	// CookieCiphertext 是加密后的 Cookie。
	CookieCiphertext string `json:"-" gorm:"type:text;not null"`
	// TokenCiphertext 是加密后的访问令牌。
	TokenCiphertext string `json:"-" gorm:"type:text"`
	// TokenExpiresAt 是访问令牌到期时间。
	TokenExpiresAt *time.Time `json:"-"`
	// Version 标识凭据格式或密钥版本。
	Version uint `json:"-" gorm:"not null;default:1"`
}

// TableName 返回账号凭据表名。
func (Credential) TableName() string { return "xy_credentials" }
