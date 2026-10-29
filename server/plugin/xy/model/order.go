package model

import "github.com/flipped-aurora/gin-vue-admin/server/global"

const (
	OrderStatusUnknown         = "unknown"
	OrderStatusPendingDelivery = "pending_delivery"
	OrderStatusShipped         = "shipped"
)

type Order struct {
	global.GVA_MODEL
	OwnerID           uint   `json:"-" gorm:"not null;index:idx_xy_orders_owner"`
	AccountID         uint   `json:"accountId" gorm:"not null;uniqueIndex:idx_xy_orders_account_external"`
	ExternalOrderID   string `json:"externalOrderId" gorm:"size:128;not null;uniqueIndex:idx_xy_orders_account_external"`
	ProductID         uint   `json:"productId" gorm:"index"`
	ExternalProductID string `json:"externalProductId" gorm:"size:128;index"`
	ProductTitle      string `json:"productTitle" gorm:"size:255"`
	BuyerUserID       string `json:"buyerUserId" gorm:"size:128"`
	BuyerUserName     string `json:"buyerUserName" gorm:"size:255"`
	SessionID         string `json:"sessionId" gorm:"size:255;index"`
	SKUKey            string `json:"skuKey" gorm:"size:128"`
	Status            string `json:"status" gorm:"size:32;not null;index"`
	AmountCent        int64  `json:"amountCent" gorm:"not null;default:0"`
}

func (Order) TableName() string { return "xy_orders" }

type OrderEvent struct {
	global.GVA_MODEL
	OwnerID   uint   `json:"-" gorm:"not null;index:idx_xy_order_events_owner"`
	AccountID uint   `json:"accountId" gorm:"not null;uniqueIndex:idx_xy_order_events_account_pnm"`
	OrderID   uint   `json:"orderId" gorm:"not null;index"`
	PNMID     string `json:"pnmId" gorm:"size:255;not null;uniqueIndex:idx_xy_order_events_account_pnm"`
	Type      string `json:"type" gorm:"size:32;not null"`
	EventTime int64  `json:"eventTime" gorm:"not null;index"`
}

func (OrderEvent) TableName() string { return "xy_order_events" }
