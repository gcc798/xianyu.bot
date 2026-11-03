package model

import (
	"time"
)

const (
	// RelistTaskPending 表示重发任务等待执行。
	RelistTaskPending = "pending"
	// RelistTaskPublishing 表示正在发布新商品。
	RelistTaskPublishing = "publishing"
	// RelistTaskFailed 表示重发明确失败。
	RelistTaskFailed = "failed"
	// RelistTaskUnknown 表示执行结果尚未确认。
	RelistTaskUnknown = "unknown"
	// RelistTaskSucceeded 表示新商品发布成功。
	RelistTaskSucceeded = "succeeded"
)

// RelistTask 记录售出商品的自动重发进度。
type RelistTask struct {
	Timestamps                     // Timestamps 保存任务记录的生命周期字段。
	OwnerID              uint      `json:"-" gorm:"not null;index:idx_xy_relist_tasks_owner"` // OwnerID 是任务所属的系统用户 ID。
	AccountID            uint      `json:"accountId" gorm:"not null;index"`                   // AccountID 是执行重发的闲鱼账号 ID。
	OrderID              uint      `json:"orderId" gorm:"not null;uniqueIndex"`               // OrderID 是触发任务的订单 ID。
	ProductID            uint      `json:"productId" gorm:"not null;index"`                   // ProductID 是待重发的本地商品 ID。
	Status               string    `json:"status" gorm:"size:32;not null;index"`              // Status 是任务执行状态。
	Attempts             int       `json:"attempts" gorm:"not null;default:0"`                // Attempts 是已执行次数。
	NextAttemptAt        time.Time `json:"nextAttemptAt" gorm:"not null;index"`               // NextAttemptAt 是下次允许执行的时间。
	NewExternalProductID string    `json:"newExternalProductId" gorm:"size:128"`              // NewExternalProductID 是成功发布后的闲鱼商品标识。
	LastError            string    `json:"lastError" gorm:"size:500"`                         // LastError 是最近一次执行失败原因。
}

// TableName 返回商品重发任务表名。
func (RelistTask) TableName() string { return "xy_relist_tasks" }
