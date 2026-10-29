package service

import "context"

var Service = newService()

func newService() *service {
	accounts := new(account)
	rules := new(replyRule)
	deliveryRules := new(deliveryRule)
	orders := &order{mtop: new(mtopClient)}
	products := new(product)
	relists := &relist{mtop: new(mtopClient), products: products}
	deliveries := &delivery{consigner: orders, rules: deliveryRules, relist: relists, workers: make(map[uint]context.CancelFunc)}
	orders.planner = deliveries
	messages := &message{rules: rules, orders: orders, delivery: deliveries, pending: make(map[string]pendingReply)}
	webSocket := newWebSocketService(messages)
	webSocket.accounts = accounts
	messages.sender = webSocket
	orders.sender = webSocket
	webSocket.delivery = deliveries
	return &service{Account: *accounts, Product: *products, Message: messages, ReplyRule: *rules, DeliveryRule: *deliveryRules, Delivery: deliveries, Relist: relists, Order: orders, QRLogin: newQRLogin(), WebSocket: webSocket}
}

type service struct {
	Status        status
	Account       account
	Product       product
	ReplyRule     replyRule
	DeliveryRule  deliveryRule
	Delivery      *delivery
	Relist        *relist
	Order         *order
	Message       *message
	QRLogin       qrLogin
	WebSocket     *webSocketService
	BusinessEvent businessEvent
	AIConfig      aiConfig
	Resource      resourceCatalog
}
