package request

import commonRequest "github.com/gcc798/xianyu.bot/internal/model/common/request"

type DeliveryRuleSearch struct {
	commonRequest.PageInfo
	ProductID uint   `json:"productId" form:"productId"`
	Mode      string `json:"mode" form:"mode"`
}

type SaveDeliveryRule struct {
	ID        uint   `json:"ID"`
	ProductID uint   `json:"productId"`
	SKUKey    string `json:"skuKey"`
	SKUName   string `json:"skuName"`
	Mode      string `json:"mode"`
	Template  string `json:"template"`
	Enabled   bool   `json:"enabled"`
}

type DeliveryRuleID struct {
	ID uint `json:"ID" form:"ID"`
}
