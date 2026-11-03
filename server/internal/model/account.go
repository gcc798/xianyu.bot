package model

import (
	"time"
)

// Account 表示归属某个系统用户的闲鱼账号。
type Account struct {
	Timestamps            // Timestamps 保存账号记录的生命周期字段。
	OwnerID        uint   `json:"-" gorm:"not null;index:idx_xy_accounts_owner"`       // OwnerID 是账号所属的系统用户 ID。
	ExternalUserID string `json:"externalUserId" gorm:"size:64;index"`                 // ExternalUserID 是闲鱼侧用户标识。
	DeviceID       string `json:"-" gorm:"size:128"`                                   // DeviceID 是登录设备标识，不向客户端返回。
	Label          string `json:"label" gorm:"size:80;not null"`                       // Label 是用户设置的账号显示名称。
	Note           string `json:"note" gorm:"size:500"`                                // Note 是账号备注。
	Status         string `json:"status" gorm:"size:32;not null;default:unconfigured"` // Status 是账号连接或配置状态。
	Enabled        bool   `json:"enabled" gorm:"not null;default:true"`                // Enabled 决定账号是否参与自动化任务。
	HasCredential  bool   `json:"hasCredential" gorm:"->;-:migration"`                 // HasCredential 表示账号是否已有凭据，仅用于查询投影。
}

// TableName 返回闲鱼账号表名。
func (Account) TableName() string { return "xy_accounts" }

// Credential 保存闲鱼账号的加密认证材料。
type Credential struct {
	Timestamps                  // Timestamps 保存凭据记录的生命周期字段。
	AccountID        uint       `json:"-" gorm:"not null;uniqueIndex"` // AccountID 是凭据对应的闲鱼账号 ID。
	CookieCiphertext string     `json:"-" gorm:"type:text;not null"`   // CookieCiphertext 是加密后的 Cookie。
	TokenCiphertext  string     `json:"-" gorm:"type:text"`            // TokenCiphertext 是加密后的访问令牌。
	TokenExpiresAt   *time.Time `json:"-"`                             // TokenExpiresAt 是访问令牌到期时间。
	Version          uint       `json:"-" gorm:"not null;default:1"`   // Version 标识凭据格式或密钥版本。
}

// TableName 返回账号凭据表名。
func (Credential) TableName() string { return "xy_credentials" }
