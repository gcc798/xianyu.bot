package request

type ImportDeliveryResources struct {
	RuleID uint     `json:"ruleId"`
	Values []string `json:"values"`
}

type DeliveryTaskSearch struct {
	PageInfo
	AccountID uint   `json:"accountId" form:"accountId"`
	Status    string `json:"status" form:"status"`
	Keyword   string `json:"keyword" form:"keyword"`
}

type DeliveryTaskID struct {
	ID uint `json:"ID" form:"ID"`
}
