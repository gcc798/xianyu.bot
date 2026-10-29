package router

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/api"

var (
	Router           = new(router)
	statusApi        = api.Api.Status
	accountApi       = api.Api.Account
	productApi       = api.Api.Product
	replyRuleApi     = api.Api.ReplyRule
	deliveryRuleApi  = api.Api.DeliveryRule
	messageApi       = api.Api.Message
	orderApi         = api.Api.Order
	qrLoginApi       = api.Api.QRLogin
	webSocketApi     = api.Api.WebSocket
	businessEventApi = api.Api.BusinessEvent
	telegramProxyApi = api.Api.TelegramProxy
)

type router struct {
	Status        status
	Account       account
	Product       product
	ReplyRule     replyRule
	DeliveryRule  deliveryRule
	Message       message
	Order         order
	QRLogin       qrLogin
	WebSocket     webSocket
	BusinessEvent businessEvent
	TelegramProxy telegramProxy
}
