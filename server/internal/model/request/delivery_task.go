package request

// ImportDeliveryResources 描述向规则批量导入卡密的输入。
type ImportDeliveryResources struct {
	// RuleID 是接收卡密的发货规则 ID。
	RuleID uint `json:"ruleId"`
	// Values 是待加密保存的卡密原文列表。
	Values []string `json:"values"`
}

// DeliveryTaskSearch 描述自动发货任务列表的筛选条件。
type DeliveryTaskSearch struct {
	// PageInfo 提供分页和关键词条件。
	PageInfo
	// AccountID 仅返回指定闲鱼账号的任务。
	AccountID uint `json:"accountId" form:"accountId"`
	// Status 仅返回指定状态的任务。
	Status string `json:"status" form:"status"`
	// Keyword 在订单与错误信息中检索。
	Keyword string `json:"keyword" form:"keyword"`
}

// DeliveryTaskID 标识一个自动发货任务。
type DeliveryTaskID struct {
	// ID 是目标任务 ID。
	ID uint `json:"ID" form:"ID"`
}
