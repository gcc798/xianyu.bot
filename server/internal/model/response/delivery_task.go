package response

import "time"

// DeliveryInventory 汇总一条发货规则的卡密库存。
type DeliveryInventory struct {
	RuleID    uint  `json:"ruleId"`    // RuleID 是发货规则 ID。
	Available int64 `json:"available"` // Available 是可领取卡密数。
	Reserved  int64 `json:"reserved"`  // Reserved 是已预留卡密数。
	Consumed  int64 `json:"consumed"`  // Consumed 是已交付卡密数。
}

// DeliveryTask 展示一条带账号、订单和商品信息的发货任务。
type DeliveryTask struct {
	ID              uint       `json:"ID"`              // ID 是发货任务 ID。
	CreatedAt       time.Time  `json:"CreatedAt"`       // CreatedAt 是任务创建时间。
	UpdatedAt       time.Time  `json:"UpdatedAt"`       // UpdatedAt 是任务最后更新时间。
	AccountID       uint       `json:"accountId"`       // AccountID 是执行发货的闲鱼账号 ID。
	AccountLabel    string     `json:"accountLabel"`    // AccountLabel 是账号显示名称。
	ExternalOrderID string     `json:"externalOrderId"` // ExternalOrderID 是闲鱼侧订单标识。
	ProductTitle    string     `json:"productTitle"`    // ProductTitle 是商品标题。
	BuyerUserName   string     `json:"buyerUserName"`   // BuyerUserName 是买家显示名称。
	Mode            string     `json:"mode"`            // Mode 是文本或卡密发货模式。
	Status          string     `json:"status"`          // Status 是任务调度或发送状态。
	Attempts        int        `json:"attempts"`        // Attempts 是已执行次数。
	NextAttemptAt   time.Time  `json:"nextAttemptAt"`   // NextAttemptAt 是下次允许执行的时间。
	LastError       string     `json:"lastError"`       // LastError 是最近一次执行失败原因。
	SentAt          *time.Time `json:"sentAt"`          // SentAt 是消息首次成功提交时间。
	ConfirmedAt     *time.Time `json:"confirmedAt"`     // ConfirmedAt 是发货确认完成时间。
}
