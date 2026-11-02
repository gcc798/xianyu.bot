package router

import api "github.com/gcc798/xianyu.bot/internal/api/xianyu"

type dependencies struct {
	handlers *api.API
}

type Router struct {
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

func New(handlers *api.API) *Router {
	deps := &dependencies{handlers: handlers}
	return &Router{
		Status: status{deps}, Account: account{deps}, Product: product{deps},
		ReplyRule: replyRule{deps}, DeliveryRule: deliveryRule{deps}, Message: message{deps},
		Order: order{deps}, QRLogin: qrLogin{deps}, WebSocket: webSocket{deps},
		BusinessEvent: businessEvent{deps}, TelegramProxy: telegramProxy{deps}, AIConfig: aiConfig{deps},
		Resource: resource{deps}, AgentSession: agentSession{deps},
	}
}
