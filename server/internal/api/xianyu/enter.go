package api

import service "github.com/gcc798/xianyu.bot/internal/service"

type dependencies struct {
	services *service.Service
}

type API struct {
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
	AIConfig      aiConfig
	Resource      resource
	AgentSession  agentSession
}

func New(services *service.Service) *API {
	deps := &dependencies{services: services}
	return &API{
		Status: status{deps}, Account: account{deps}, Product: product{deps},
		ReplyRule: replyRule{deps}, DeliveryRule: deliveryRule{deps}, Message: message{deps},
		Order: order{deps}, QRLogin: qrLogin{deps}, WebSocket: webSocket{deps},
		BusinessEvent: businessEvent{deps}, TelegramProxy: telegramProxy{}, AIConfig: aiConfig{deps},
		Resource: resource{deps}, AgentSession: agentSession{deps},
	}
}
