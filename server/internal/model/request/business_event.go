package request

// BusinessEventSearch 描述业务事件列表的筛选条件。
type BusinessEventSearch struct {
	PageInfo         // PageInfo 提供分页和关键词条件。
	AccountID uint   `form:"accountId"` // AccountID 仅返回指定账号的事件。
	ProductID uint   `form:"productId"` // ProductID 仅返回指定商品的事件。
	OrderID   uint   `form:"orderId"`   // OrderID 仅返回指定订单的事件。
	Type      string `form:"type"`      // Type 仅返回指定业务类别的事件。
	Status    string `form:"status"`    // Status 仅返回指定结果状态的事件。
	Keyword   string `form:"keyword"`   // Keyword 在事件摘要和错误信息中检索。
}
