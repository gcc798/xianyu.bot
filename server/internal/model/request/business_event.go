package request

type BusinessEventSearch struct {
	PageInfo
	AccountID uint   `form:"accountId"`
	ProductID uint   `form:"productId"`
	OrderID   uint   `form:"orderId"`
	Type      string `form:"type"`
	Status    string `form:"status"`
	Keyword   string `form:"keyword"`
}
