package request

// ImportDeliveryResources 描述向规则批量导入卡密的输入。
type ImportDeliveryResources struct {
	RuleID uint     `json:"ruleId"` // RuleID 是接收卡密的发货规则 ID。
	Values []string `json:"values"` // Values 是待加密保存的卡密原文列表。
}

// DeliveryTaskSearch 描述自动发货任务列表的筛选条件。
type DeliveryTaskSearch struct {
	PageInfo         // PageInfo 提供分页和关键词条件。
	AccountID uint   `json:"accountId" form:"accountId"` // AccountID 仅返回指定闲鱼账号的任务。
	ProductID uint   `json:"productId" form:"productId"` // ProductID 仅返回指定商品的任务。
	Status    string `json:"status" form:"status"`       // Status 仅返回指定状态的任务。
	Keyword   string `json:"keyword" form:"keyword"`     // Keyword 在订单与错误信息中检索。
}

// DeliveryTaskID 标识一个自动发货任务。
type DeliveryTaskID struct {
	ID uint `json:"ID" form:"ID"` // ID 是目标任务 ID。
}
