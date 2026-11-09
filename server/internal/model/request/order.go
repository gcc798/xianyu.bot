package request

// OrderSearch 描述订单列表的分页与筛选条件。
type OrderSearch struct {
	PageInfo         // PageInfo 提供分页和关键词条件。
	AccountID uint   `json:"accountId" form:"accountId"` // AccountID 仅返回指定闲鱼账号的订单。
	Status    string `json:"status" form:"status"`       // Status 仅返回指定状态的订单。
}

// OrderSync 描述手动同步订单的目标账号。
type OrderSync struct {
	AccountID uint `json:"accountId"` // AccountID 是待同步的闲鱼账号 ID。
}
