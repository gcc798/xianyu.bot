package response

import "time"

// BusinessEvent 展示一条带关联对象名称的业务审计事件。
type BusinessEvent struct {
	// ID 是事件记录 ID。
	ID uint `json:"ID"`
	// CreatedAt 是事件创建时间。
	CreatedAt time.Time `json:"CreatedAt"`
	// AccountID 是关联的闲鱼账号 ID。
	AccountID uint `json:"accountId"`
	// AccountLabel 是账号显示名称。
	AccountLabel string `json:"accountLabel"`
	// ProductID 是关联的本地商品 ID。
	ProductID uint `json:"productId"`
	// ProductTitle 是商品标题。
	ProductTitle string `json:"productTitle"`
	// OrderID 是关联的本地订单 ID。
	OrderID uint `json:"orderId"`
	// ExternalOrderID 是闲鱼侧订单标识。
	ExternalOrderID string `json:"externalOrderId"`
	// Type 是业务事件类别。
	Type string `json:"type"`
	// Status 是业务动作结果状态。
	Status string `json:"status"`
	// Summary 是供用户查看的事件摘要。
	Summary string `json:"summary"`
	// Error 是失败时的安全错误描述。
	Error string `json:"error"`
}
