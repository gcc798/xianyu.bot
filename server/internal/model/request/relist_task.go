package request

// RelistTaskSearch 描述自动续售任务列表的筛选条件。
type RelistTaskSearch struct {
	Pagination        // Pagination 提供页码与每页数量。
	AccountID  uint   `json:"accountId"` // AccountID 仅返回指定账号的续售任务。
	ProductID  uint   `json:"productId"` // ProductID 仅返回指定商品的续售任务。
	Status     string `json:"status"`    // Status 仅返回指定状态的续售任务。
}
