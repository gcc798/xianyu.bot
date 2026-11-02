package request

type DeliveryRuleSearch struct {
	PageInfo
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
