package response

import "time"

// BusinessEvent 展示一条带关联对象名称的业务审计事件。
type BusinessEvent struct {
	ID              uint      `json:"ID"`              // ID 是事件记录 ID。
	CreatedAt       time.Time `json:"CreatedAt"`       // CreatedAt 是事件创建时间。
	AccountID       uint      `json:"accountId"`       // AccountID 是关联的闲鱼账号 ID。
	AccountLabel    string    `json:"accountLabel"`    // AccountLabel 是账号显示名称。
	ProductID       uint      `json:"productId"`       // ProductID 是关联的本地商品 ID。
	ProductTitle    string    `json:"productTitle"`    // ProductTitle 是商品标题。
	OrderID         uint      `json:"orderId"`         // OrderID 是关联的本地订单 ID。
	ExternalOrderID string    `json:"externalOrderId"` // ExternalOrderID 是闲鱼侧订单标识。
	Type            string    `json:"type"`            // Type 是业务事件类别。
	Status          string    `json:"status"`          // Status 是业务动作结果状态。
	Summary         string    `json:"summary"`         // Summary 是供用户查看的事件摘要。
	Error           string    `json:"error"`           // Error 是失败时的安全错误描述。
}
