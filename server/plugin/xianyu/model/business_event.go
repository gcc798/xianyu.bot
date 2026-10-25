package model

import "github.com/flipped-aurora/gin-vue-admin/server/global"

type BusinessEvent struct {
	global.GVA_MODEL
	OwnerID   uint   `json:"-" gorm:"not null;index:idx_xy_business_events_owner"`
	AccountID uint   `json:"accountId" gorm:"index"`
	ProductID uint   `json:"productId" gorm:"index"`
	OrderID   uint   `json:"orderId" gorm:"index"`
	Type      string `json:"type" gorm:"size:32;not null;index"`
	Status    string `json:"status" gorm:"size:16;not null;index"`
	Summary   string `json:"summary" gorm:"size:255;not null"`
	Error     string `json:"error" gorm:"size:500"`
}

func (BusinessEvent) TableName() string { return "xy_business_events" }
