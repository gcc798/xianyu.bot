package service

import "context"

import "gorm.io/gorm"

type Dependencies struct {
	DB                    *gorm.DB
	SendNotificationEmail func(string, string, string) error
}

func New(deps Dependencies) *Services {
	accounts := &account{db: deps.DB}
	events := &businessEvent{db: deps.DB}
	rules := &replyRule{db: deps.DB}
	deliveryRules := &deliveryRule{db: deps.DB}
	orders := &order{db: deps.DB, events: events, mtop: &mtopClient{db: deps.DB}}
	products := &product{db: deps.DB, events: events}
	relists := &relist{db: deps.DB, events: events, mtop: &mtopClient{db: deps.DB}, products: products}
	deliveries := &delivery{db: deps.DB, events: events, consigner: orders, rules: deliveryRules, relist: relists, workers: make(map[uint]context.CancelFunc)}
	orders.planner = deliveries
	messages := &message{rules: rules, orders: orders, delivery: deliveries, pending: make(map[string]pendingReply)}
	webSocket := newWebSocketService(messages, deps.SendNotificationEmail)
	webSocket.accounts = accounts
	accounts.webSocket = webSocket
	messages.sender = webSocket
	orders.sender = webSocket
	webSocket.delivery = deliveries
	agentEvents := newAgentEventHub()
	agentRunner := &agentRunner{events: agentEvents, workers: make(map[uint]context.CancelFunc)}
	return &Services{Status: status{db: deps.DB}, Account: *accounts, Product: *products, Message: messages, ReplyRule: *rules, DeliveryRule: *deliveryRules, Delivery: deliveries, Relist: relists, Order: orders, QRLogin: newQRLogin(accounts), WebSocket: webSocket, BusinessEvent: *events, AIConfig: aiConfig{db: deps.DB}, Resource: resourceCatalog{events: agentEvents}, AgentSession: agentSession{db: deps.DB}, AgentRunner: agentRunner, AgentEvents: agentEvents}
}

type Services struct {
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
	AgentSession  agentSession
	AgentRunner   *agentRunner
	AgentEvents   *agentEventHub
}
