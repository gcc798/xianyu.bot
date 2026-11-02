package request

import commonRequest "github.com/gcc798/go.xianyu.assistant/server/internal/model/common/request"

type ImportDeliveryResources struct {
	RuleID uint     `json:"ruleId"`
	Values []string `json:"values"`
}

type DeliveryTaskSearch struct {
	commonRequest.PageInfo
	AccountID uint   `json:"accountId" form:"accountId"`
	Status    string `json:"status" form:"status"`
	Keyword   string `json:"keyword" form:"keyword"`
}

type DeliveryTaskID struct {
	ID uint `json:"ID" form:"ID"`
}
