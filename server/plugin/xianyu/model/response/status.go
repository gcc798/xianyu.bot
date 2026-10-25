package response

type Status struct {
	Plugin           string `json:"plugin"`
	Database         string `json:"database"`
	Accounts         int64  `json:"accounts"`
	ReadyAccounts    int64  `json:"readyAccounts"`
	OnSaleProducts   int64  `json:"onSaleProducts"`
	Messages         int64  `json:"messages"`
	PendingOrders    int64  `json:"pendingOrders"`
	FailedDeliveries int64  `json:"failedDeliveries"`
}
