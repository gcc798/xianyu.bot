package model

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

type Account struct {
	global.GVA_MODEL
	OwnerID        uint   `json:"-" gorm:"not null;index:idx_xy_accounts_owner"`
	ExternalUserID string `json:"externalUserId" gorm:"size:64;index"`
	DeviceID       string `json:"-" gorm:"size:128"`
	Label          string `json:"label" gorm:"size:80;not null"`
	Note           string `json:"note" gorm:"size:500"`
	Status         string `json:"status" gorm:"size:32;not null;default:unconfigured"`
	Enabled        bool   `json:"enabled" gorm:"not null;default:true"`
	HasCredential  bool   `json:"hasCredential" gorm:"->;-:migration"`
}

func (Account) TableName() string { return "xy_accounts" }

type Credential struct {
	global.GVA_MODEL
	AccountID        uint       `json:"-" gorm:"not null;uniqueIndex"`
	CookieCiphertext string     `json:"-" gorm:"type:text;not null"`
	TokenCiphertext  string     `json:"-" gorm:"type:text"`
	TokenExpiresAt   *time.Time `json:"-"`
	Version          uint       `json:"-" gorm:"not null;default:1"`
}

func (Credential) TableName() string { return "xy_credentials" }
