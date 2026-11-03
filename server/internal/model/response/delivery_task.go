package response

import "time"

// DeliveryInventory 汇总一条发货规则的卡密库存。
type DeliveryInventory struct {
	// RuleID 是发货规则 ID。
	RuleID uint `json:"ruleId"`
	// Available 是可领取卡密数。
	Available int64 `json:"available"`
	// Reserved 是已预留卡密数。
	Reserved int64 `json:"reserved"`
	// Consumed 是已交付卡密数。
	Consumed int64 `json:"consumed"`
}

// DeliveryTask 展示一条带账号、订单和商品信息的发货任务。
type DeliveryTask struct {
	// ID 是发货任务 ID。
	ID uint `json:"ID"`
	// CreatedAt 是任务创建时间。
	CreatedAt time.Time `json:"CreatedAt"`
	// UpdatedAt 是任务最后更新时间。
	UpdatedAt time.Time `json:"UpdatedAt"`
	// AccountID 是执行发货的闲鱼账号 ID。
	AccountID uint `json:"accountId"`
	// AccountLabel 是账号显示名称。
	AccountLabel string `json:"accountLabel"`
	// ExternalOrderID 是闲鱼侧订单标识。
	ExternalOrderID string `json:"externalOrderId"`
	// ProductTitle 是商品标题。
	ProductTitle string `json:"productTitle"`
	// BuyerUserName 是买家显示名称。
	BuyerUserName string `json:"buyerUserName"`
	// Mode 是文本或卡密发货模式。
	Mode string `json:"mode"`
	// Status 是任务调度或发送状态。
	Status string `json:"status"`
	// Attempts 是已执行次数。
	Attempts int `json:"attempts"`
	// NextAttemptAt 是下次允许执行的时间。
	NextAttemptAt time.Time `json:"nextAttemptAt"`
	// LastError 是最近一次执行失败原因。
	LastError string `json:"lastError"`
	// SentAt 是消息首次成功提交时间。
	SentAt *time.Time `json:"sentAt"`
	// ConfirmedAt 是发货确认完成时间。
	ConfirmedAt *time.Time `json:"confirmedAt"`
}
