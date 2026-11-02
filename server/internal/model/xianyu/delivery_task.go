package model

import (
	"time"

	"github.com/gcc798/xianyu.bot/internal/model/common"
)

const (
	DeliveryResourceAvailable = "available"
	DeliveryResourceReserved  = "reserved"
	DeliveryResourceConsumed  = "consumed"

	DeliveryTaskPending         = "pending"
	DeliveryTaskSending         = "sending"
	DeliveryTaskSentUnconfirmed = "sent_unconfirmed"
	DeliveryTaskFailed          = "failed"
	DeliveryTaskUnknown         = "unknown"
	DeliveryTaskConfirmed       = "confirmed"
)

type DeliveryResource struct {
	common.Timestamps
	OwnerID         uint       `json:"-" gorm:"not null;index:idx_xy_delivery_resources_owner"`
	RuleID          uint       `json:"ruleId" gorm:"not null;uniqueIndex:idx_xy_delivery_resource_rule_hash"`
	ValueHash       string     `json:"-" gorm:"size:64;not null;uniqueIndex:idx_xy_delivery_resource_rule_hash"`
	ValueCiphertext string     `json:"-" gorm:"type:text;not null"`
	Status          string     `json:"status" gorm:"size:16;not null;index"`
	ReservedTaskID  uint       `json:"reservedTaskId" gorm:"index"`
	ConsumedAt      *time.Time `json:"consumedAt"`
}

func (DeliveryResource) TableName() string { return "xy_delivery_resources" }

type DeliveryTask struct {
	common.Timestamps
	OwnerID           uint       `json:"-" gorm:"not null;index:idx_xy_delivery_tasks_owner"`
	AccountID         uint       `json:"accountId" gorm:"not null;index"`
	OrderID           uint       `json:"orderId" gorm:"not null;uniqueIndex"`
	RuleID            uint       `json:"ruleId" gorm:"not null"`
	ResourceID        uint       `json:"resourceId" gorm:"index"`
	PayloadCiphertext string     `json:"-" gorm:"type:text;not null"`
	Status            string     `json:"status" gorm:"size:32;not null;index"`
	Attempts          int        `json:"attempts" gorm:"not null;default:0"`
	NextAttemptAt     time.Time  `json:"nextAttemptAt" gorm:"not null;index"`
	LastError         string     `json:"lastError" gorm:"size:500"`
	SentAt            *time.Time `json:"sentAt"`
	ConfirmedAt       *time.Time `json:"confirmedAt"`
}

func (DeliveryTask) TableName() string { return "xy_delivery_tasks" }
