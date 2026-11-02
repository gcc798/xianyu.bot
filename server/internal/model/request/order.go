package request

type OrderSearch struct {
	PageInfo
	AccountID uint   `json:"accountId" form:"accountId"`
	Status    string `json:"status" form:"status"`
}

type OrderSync struct {
	AccountID uint `json:"accountId"`
}
