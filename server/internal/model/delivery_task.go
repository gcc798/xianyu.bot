package model

import (
	"time"
)

const (
	// DeliveryResourceAvailable 表示卡密可被新任务领取。
	DeliveryResourceAvailable = "available"
	// DeliveryResourceReserved 表示卡密已被发货任务预留。
	DeliveryResourceReserved = "reserved"
	// DeliveryResourceConsumed 表示卡密已完成交付。
	DeliveryResourceConsumed = "consumed"

	// DeliveryTaskPending 表示发货任务等待执行。
	DeliveryTaskPending = "pending"
	// DeliveryTaskSending 表示发货消息正在发送。
	DeliveryTaskSending = "sending"
	// DeliveryTaskSentUnconfirmed 表示消息已发送但结果未确认。
	DeliveryTaskSentUnconfirmed = "sent_unconfirmed"
	// DeliveryTaskFailed 表示发货明确失败。
	DeliveryTaskFailed = "failed"
	// DeliveryTaskUnknown 表示发货结果未知。
	DeliveryTaskUnknown = "unknown"
	// DeliveryTaskConfirmed 表示买家消息已确认送达。
	DeliveryTaskConfirmed = "confirmed"
)

// DeliveryResource 保存加密卡密及其占用状态。
type DeliveryResource struct {
	Timestamps                 // Timestamps 保存卡密记录的生命周期字段。
	OwnerID         uint       `json:"-" gorm:"not null;index:idx_xy_delivery_resources_owner"`                  // OwnerID 是卡密所属的系统用户 ID。
	RuleID          uint       `json:"ruleId" gorm:"not null;uniqueIndex:idx_xy_delivery_resource_rule_hash"`    // RuleID 是卡密所属的发货规则 ID。
	ValueHash       string     `json:"-" gorm:"size:64;not null;uniqueIndex:idx_xy_delivery_resource_rule_hash"` // ValueHash 用于在同一规则内去重卡密。
	ValueCiphertext string     `json:"-" gorm:"type:text;not null"`                                              // ValueCiphertext 是加密后的卡密原文。
	Status          string     `json:"status" gorm:"size:16;not null;index"`                                     // Status 是可用、预留或已消费状态。
	ReservedTaskID  uint       `json:"reservedTaskId" gorm:"index"`                                              // ReservedTaskID 是当前占用卡密的发货任务 ID。
	ConsumedAt      *time.Time `json:"consumedAt"`                                                               // ConsumedAt 是卡密确认交付时间。
}

// TableName 返回发货资源表名。
func (DeliveryResource) TableName() string { return "xy_delivery_resources" }

// DeliveryTask 记录订单自动发货的执行状态。
type DeliveryTask struct {
	Timestamps                   // Timestamps 保存任务记录的生命周期字段。
	OwnerID           uint       `json:"-" gorm:"not null;index:idx_xy_delivery_tasks_owner"` // OwnerID 是任务所属的系统用户 ID。
	AccountID         uint       `json:"accountId" gorm:"not null;index"`                     // AccountID 是执行发货的闲鱼账号 ID。
	OrderID           uint       `json:"orderId" gorm:"not null;uniqueIndex"`                 // OrderID 是待发货的本地订单 ID。
	RuleID            uint       `json:"ruleId" gorm:"not null"`                              // RuleID 是本次发货采用的规则 ID。
	ResourceID        uint       `json:"resourceId" gorm:"index"`                             // ResourceID 是本次发货预留的卡密 ID。
	PayloadCiphertext string     `json:"-" gorm:"type:text;not null"`                         // PayloadCiphertext 是加密后的最终发货消息。
	Status            string     `json:"status" gorm:"size:32;not null;index"`                // Status 是任务调度或发送状态。
	Attempts          int        `json:"attempts" gorm:"not null;default:0"`                  // Attempts 是已执行次数。
	NextAttemptAt     time.Time  `json:"nextAttemptAt" gorm:"not null;index"`                 // NextAttemptAt 是下次允许执行的时间。
	LastError         string     `json:"lastError" gorm:"size:500"`                           // LastError 是最近一次执行失败原因。
	SentAt            *time.Time `json:"sentAt"`                                              // SentAt 是消息首次成功提交时间。
	ConfirmedAt       *time.Time `json:"confirmedAt"`                                         // ConfirmedAt 是发货确认完成时间。
}

// TableName 返回自动发货任务表名。
func (DeliveryTask) TableName() string { return "xy_delivery_tasks" }
