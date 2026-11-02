package model

import (
	"time"

	"github.com/gcc798/go.xianyu.assistant/server/internal/model/common"
)

const (
	RelistTaskPending    = "pending"
	RelistTaskPublishing = "publishing"
	RelistTaskFailed     = "failed"
	RelistTaskUnknown    = "unknown"
	RelistTaskSucceeded  = "succeeded"
)

type RelistTask struct {
	common.Timestamps
	OwnerID              uint      `json:"-" gorm:"not null;index:idx_xy_relist_tasks_owner"`
	AccountID            uint      `json:"accountId" gorm:"not null;index"`
	OrderID              uint      `json:"orderId" gorm:"not null;uniqueIndex"`
	ProductID            uint      `json:"productId" gorm:"not null;index"`
	Status               string    `json:"status" gorm:"size:32;not null;index"`
	Attempts             int       `json:"attempts" gorm:"not null;default:0"`
	NextAttemptAt        time.Time `json:"nextAttemptAt" gorm:"not null;index"`
	NewExternalProductID string    `json:"newExternalProductId" gorm:"size:128"`
	LastError            string    `json:"lastError" gorm:"size:500"`
}

func (RelistTask) TableName() string { return "xy_relist_tasks" }
