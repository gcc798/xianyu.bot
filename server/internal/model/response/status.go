package response

// Status 汇总服务健康状态与关键业务计数。
type Status struct {
	Database         string `json:"database"`         // Database 是数据库连接状态。
	Accounts         int64  `json:"accounts"`         // Accounts 是闲鱼账号总数。
	ReadyAccounts    int64  `json:"readyAccounts"`    // ReadyAccounts 是可用闲鱼账号数。
	OnSaleProducts   int64  `json:"onSaleProducts"`   // OnSaleProducts 是在售商品数。
	Messages         int64  `json:"messages"`         // Messages 是消息总数。
	PendingOrders    int64  `json:"pendingOrders"`    // PendingOrders 是待发货订单数。
	FailedDeliveries int64  `json:"failedDeliveries"` // FailedDeliveries 是失败发货任务数。
}
