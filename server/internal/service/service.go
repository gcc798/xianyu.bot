package service

import (
	"context"

	"github.com/gcc798/xianyu.bot/internal/agent"
	"gorm.io/gorm"
)

type Service struct {
	DB               *gorm.DB
	RevokedTokens    *RevokedTokenService
	User             *UserService
	Dictionary       *DictionaryService
	DictionaryDetail *DictionaryDetailService
	LoginLog         *LoginLogService
	Status           *Status
	Account          *Account
	Product          *Product
	ReplyRule        *ReplyRule
	DeliveryRule     *DeliveryRule
	Delivery         *Delivery
	Relist           *Relist
	Order            *Order
	Message          *Message
	QRLogin          *QRLogin
	WebSocket        *WebSocketService
	BusinessEvent    *BusinessEvent
	AIConfig         *agent.AIConfig
	Resource         *agent.ResourceCatalog
	AgentSession     *agent.Session
	AgentRunner      *agent.Runner
	AgentEvents      *agent.EventHub
}

func New(ctx context.Context, db *gorm.DB, sendNotificationEmail func(string, string, string) error) *Service {
	root := contextOrBackground(ctx)
	events := NewBusinessEvent(db)
	accounts := NewAccount(db, events)
	rules := NewReplyRule(db)
	deliveryRules := NewDeliveryRule(db)
	orders := NewOrder(db, events)
	products := NewProduct(db, events)
	relists := NewRelist(db, events, products)
	deliveries := NewDelivery(root, db, events, orders, deliveryRules, relists)
	orders.planner = deliveries
	messages := NewMessage(root, db, events, rules, orders, deliveries)
	webSocket := NewWebSocketService(root, db, events, messages, sendNotificationEmail)
	webSocket.accounts = accounts
	accounts.webSocket = webSocket
	messages.sender = webSocket
	orders.sender = webSocket
	webSocket.delivery = deliveries
	agentEvents := agent.NewEventHub(db)
	mcpRegistry := agent.NewMCPRegistry(db)
	resources := agent.NewResourceCatalog(db, agentEvents, mcpRegistry)
	return &Service{
		DB:            db,
		RevokedTokens: NewRevokedTokenService(db), User: NewUserService(db),
		Dictionary: NewDictionaryService(db), DictionaryDetail: NewDictionaryDetailService(db), LoginLog: NewLoginLogService(db),
		Status: NewStatus(db), Account: accounts, Product: products, ReplyRule: rules, DeliveryRule: deliveryRules,
		Delivery: deliveries, Relist: relists, Order: orders, Message: messages,
		QRLogin: NewQRLogin(root, accounts), WebSocket: webSocket, BusinessEvent: events,
		AIConfig: agent.NewAIConfig(db), Resource: resources, AgentSession: agent.NewSession(db),
		AgentRunner: agent.NewRunner(root, db, mcpRegistry, resources, agentEvents), AgentEvents: agentEvents,
	}
}

func contextOrBackground(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
}
