package service

import (
	"context"

	"github.com/gcc798/xianyu.bot/internal/agent"
	"gorm.io/gorm"
)

type Dependencies struct {
	Context               context.Context
	DB                    *gorm.DB
	SendNotificationEmail func(string, string, string) error
}

func New(deps Dependencies) *Services {
	root := contextOrBackground(deps.Context)
	events := &businessEvent{db: deps.DB}
	accounts := &account{db: deps.DB, events: events}
	rules := &replyRule{db: deps.DB}
	deliveryRules := &deliveryRule{db: deps.DB}
	orders := &order{db: deps.DB, events: events, mtop: &mtopClient{db: deps.DB}}
	products := &product{db: deps.DB, events: events}
	relists := &relist{db: deps.DB, events: events, mtop: &mtopClient{db: deps.DB}, products: products}
	deliveries := &delivery{root: root, db: deps.DB, events: events, consigner: orders, rules: deliveryRules, relist: relists, workers: make(map[uint]context.CancelFunc)}
	orders.planner = deliveries
	messages := &message{root: root, db: deps.DB, events: events, rules: rules, orders: orders, delivery: deliveries, pending: make(map[string]pendingReply)}
	webSocket := newWebSocketService(root, deps.DB, events, messages, deps.SendNotificationEmail)
	webSocket.accounts = accounts
	accounts.webSocket = webSocket
	messages.sender = webSocket
	orders.sender = webSocket
	webSocket.delivery = deliveries
	agentEvents := agent.NewEventHub(deps.DB)
	mcpRegistry := agent.NewMCPRegistry(deps.DB)
	resources := agent.NewResourceCatalog(deps.DB, agentEvents, mcpRegistry)
	agentRunner := agent.NewRunner(root, deps.DB, mcpRegistry, resources, agentEvents)
	return &Services{Status: status{db: deps.DB}, Account: *accounts, Product: *products, Message: messages, ReplyRule: *rules, DeliveryRule: *deliveryRules, Delivery: deliveries, Relist: relists, Order: orders, QRLogin: newQRLogin(root, accounts), WebSocket: webSocket, BusinessEvent: *events, AIConfig: agent.NewAIConfig(deps.DB), Resource: resources, AgentSession: agent.NewSession(deps.DB), AgentRunner: agentRunner, AgentEvents: agentEvents}
}

func contextOrBackground(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
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
	AIConfig      *agent.AIConfig
	Resource      *agent.ResourceCatalog
	AgentSession  *agent.Session
	AgentRunner   *agent.Runner
	AgentEvents   *agent.EventHub
}
