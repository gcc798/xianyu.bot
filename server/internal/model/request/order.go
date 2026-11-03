package request

// OrderSearch 描述订单列表的分页与筛选条件。
type OrderSearch struct {
	// PageInfo 提供分页和关键词条件。
	PageInfo
	// AccountID 仅返回指定闲鱼账号的订单。
	AccountID uint `json:"accountId" form:"accountId"`
	// Status 仅返回指定状态的订单。
	Status string `json:"status" form:"status"`
}

// OrderSync 描述手动同步订单的目标账号。
type OrderSync struct {
	// AccountID 是待同步的闲鱼账号 ID。
	AccountID uint `json:"accountId"`
}
