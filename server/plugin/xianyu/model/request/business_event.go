package request

import commonRequest "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type BusinessEventSearch struct {
	commonRequest.PageInfo
	AccountID uint   `form:"accountId"`
	ProductID uint   `form:"productId"`
	OrderID   uint   `form:"orderId"`
	Type      string `form:"type"`
	Status    string `form:"status"`
	Keyword   string `form:"keyword"`
}
