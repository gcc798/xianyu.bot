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
	Timestamps               // Timestamps 保存订单记录的生命周期字段。
	OwnerID           uint   `json:"-" gorm:"not null;index:idx_xy_orders_owner"`                                         // OwnerID 是订单所属的系统用户 ID。
	AccountID         uint   `json:"accountId" gorm:"not null;uniqueIndex:idx_xy_orders_account_external"`                // AccountID 是订单所属的闲鱼账号 ID。
	ExternalOrderID   string `json:"externalOrderId" gorm:"size:128;not null;uniqueIndex:idx_xy_orders_account_external"` // ExternalOrderID 是闲鱼侧订单标识。
	ProductID         uint   `json:"productId" gorm:"index"`                                                              // ProductID 是本地商品 ID。
	ExternalProductID string `json:"externalProductId" gorm:"size:128;index"`                                             // ExternalProductID 是闲鱼侧商品标识。
	ProductTitle      string `json:"productTitle" gorm:"size:255"`                                                        // ProductTitle 是下单时的商品标题快照。
	BuyerUserID       string `json:"buyerUserId" gorm:"size:128"`                                                         // BuyerUserID 是闲鱼买家标识。
	BuyerUserName     string `json:"buyerUserName" gorm:"size:255"`                                                       // BuyerUserName 是闲鱼买家展示名称。
	SessionID         string `json:"sessionId" gorm:"size:255;index"`                                                     // SessionID 是订单关联的会话标识。
	SKUKey            string `json:"skuKey" gorm:"size:128"`                                                              // SKUKey 是成交商品规格标识。
	Status            string `json:"status" gorm:"size:32;not null;index"`                                                // Status 是归一化后的订单状态。
	AmountCent        int64  `json:"amountCent" gorm:"not null;default:0"`                                                // AmountCent 是以分为单位的订单金额。
}

// TableName 返回闲鱼订单表名。
func (Order) TableName() string { return "xy_orders" }

// OrderEvent 保存驱动订单状态变化的原始事件索引。
type OrderEvent struct {
	Timestamps        // Timestamps 保存事件记录的生命周期字段。
	OwnerID    uint   `json:"-" gorm:"not null;index:idx_xy_order_events_owner"`                          // OwnerID 是事件所属的系统用户 ID。
	AccountID  uint   `json:"accountId" gorm:"not null;uniqueIndex:idx_xy_order_events_account_pnm"`      // AccountID 是事件所属的闲鱼账号 ID。
	OrderID    uint   `json:"orderId" gorm:"not null;index"`                                              // OrderID 是关联的本地订单 ID。
	PNMID      string `json:"pnmId" gorm:"size:255;not null;uniqueIndex:idx_xy_order_events_account_pnm"` // PNMID 是闲鱼消息事件的去重标识。
	Type       string `json:"type" gorm:"size:32;not null"`                                               // Type 是订单事件类型。
	EventTime  int64  `json:"eventTime" gorm:"not null;index"`                                            // EventTime 是闲鱼侧事件时间戳。
}

// TableName 返回订单事件表名。
func (OrderEvent) TableName() string { return "xy_order_events" }
