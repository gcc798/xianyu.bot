package api

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/xianyu/service"

var (
	Api                  = new(api)
	statusService        = service.Service.Status
	accountService       = service.Service.Account
	productService       = service.Service.Product
	replyRuleService     = service.Service.ReplyRule
	deliveryRuleService  = service.Service.DeliveryRule
	deliveryService      = service.Service.Delivery
	messageService       = service.Service.Message
	orderService         = service.Service.Order
	qrLoginService       = service.Service.QRLogin
	webSocketService     = service.Service.WebSocket
	businessEventService = service.Service.BusinessEvent
)

type api struct {
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
