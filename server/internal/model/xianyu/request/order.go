package request

import commonRequest "github.com/gcc798/xianyu.bot/internal/model/common/request"

type OrderSearch struct {
	commonRequest.PageInfo
	AccountID uint   `json:"accountId" form:"accountId"`
	Status    string `json:"status" form:"status"`
}

type OrderSync struct {
	AccountID uint `json:"accountId"`
}
