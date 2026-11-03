package model

const (
	// OrderStatusUnknown 表示尚未识别订单状态。
	OrderStatusUnknown = "unknown"
	// OrderStatusPendingDelivery 表示订单等待自动发货。
	OrderStatusPendingDelivery = "pending_delivery"
	// OrderStatusShipped 表示订单已经发货。
	OrderStatusShipped = "shipped"
)

// Order 保存从闲鱼同步的订单快照。
type Order struct {
	// Timestamps 保存订单记录的生命周期字段。
	Timestamps
	// OwnerID 是订单所属的系统用户 ID。
	OwnerID uint `json:"-" gorm:"not null;index:idx_xy_orders_owner"`
	// AccountID 是订单所属的闲鱼账号 ID。
	AccountID uint `json:"accountId" gorm:"not null;uniqueIndex:idx_xy_orders_account_external"`
	// ExternalOrderID 是闲鱼侧订单标识。
	ExternalOrderID string `json:"externalOrderId" gorm:"size:128;not null;uniqueIndex:idx_xy_orders_account_external"`
	// ProductID 是本地商品 ID。
	ProductID uint `json:"productId" gorm:"index"`
	// ExternalProductID 是闲鱼侧商品标识。
	ExternalProductID string `json:"externalProductId" gorm:"size:128;index"`
	// ProductTitle 是下单时的商品标题快照。
	ProductTitle string `json:"productTitle" gorm:"size:255"`
	// BuyerUserID 是闲鱼买家标识。
	BuyerUserID string `json:"buyerUserId" gorm:"size:128"`
	// BuyerUserName 是闲鱼买家展示名称。
	BuyerUserName string `json:"buyerUserName" gorm:"size:255"`
	// SessionID 是订单关联的会话标识。
	SessionID string `json:"sessionId" gorm:"size:255;index"`
	// SKUKey 是成交商品规格标识。
	SKUKey string `json:"skuKey" gorm:"size:128"`
	// Status 是归一化后的订单状态。
	Status string `json:"status" gorm:"size:32;not null;index"`
	// AmountCent 是以分为单位的订单金额。
	AmountCent int64 `json:"amountCent" gorm:"not null;default:0"`
}

// TableName 返回闲鱼订单表名。
func (Order) TableName() string { return "xy_orders" }

// OrderEvent 保存驱动订单状态变化的原始事件索引。
type OrderEvent struct {
	// Timestamps 保存事件记录的生命周期字段。
	Timestamps
	// OwnerID 是事件所属的系统用户 ID。
	OwnerID uint `json:"-" gorm:"not null;index:idx_xy_order_events_owner"`
	// AccountID 是事件所属的闲鱼账号 ID。
	AccountID uint `json:"accountId" gorm:"not null;uniqueIndex:idx_xy_order_events_account_pnm"`
	// OrderID 是关联的本地订单 ID。
	OrderID uint `json:"orderId" gorm:"not null;index"`
	// PNMID 是闲鱼消息事件的去重标识。
	PNMID string `json:"pnmId" gorm:"size:255;not null;uniqueIndex:idx_xy_order_events_account_pnm"`
	// Type 是订单事件类型。
	Type string `json:"type" gorm:"size:32;not null"`
	// EventTime 是闲鱼侧事件时间戳。
	EventTime int64 `json:"eventTime" gorm:"not null;index"`
}

// TableName 返回订单事件表名。
func (OrderEvent) TableName() string { return "xy_order_events" }
