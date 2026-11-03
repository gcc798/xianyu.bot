package response

// Status 汇总服务健康状态与关键业务计数。
type Status struct {
	// Database 是数据库连接状态。
	Database string `json:"database"`
	// Accounts 是闲鱼账号总数。
	Accounts int64 `json:"accounts"`
	// ReadyAccounts 是可用闲鱼账号数。
	ReadyAccounts int64 `json:"readyAccounts"`
	// OnSaleProducts 是在售商品数。
	OnSaleProducts int64 `json:"onSaleProducts"`
	// Messages 是消息总数。
	Messages int64 `json:"messages"`
	// PendingOrders 是待发货订单数。
	PendingOrders int64 `json:"pendingOrders"`
	// FailedDeliveries 是失败发货任务数。
	FailedDeliveries int64 `json:"failedDeliveries"`
}
