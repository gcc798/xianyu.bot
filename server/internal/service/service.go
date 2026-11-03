package service

import (
	"context"

	"gorm.io/gorm"
)

type Service struct {
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
	return &Service{
		RevokedTokens: NewRevokedTokenService(db), User: NewUserService(db),
		Dictionary: NewDictionaryService(db), DictionaryDetail: NewDictionaryDetailService(db), LoginLog: NewLoginLogService(db),
		Status: NewStatus(db), Account: accounts, Product: products, ReplyRule: rules, DeliveryRule: deliveryRules,
		Delivery: deliveries, Relist: relists, Order: orders, Message: messages,
		QRLogin: NewQRLogin(root, accounts), WebSocket: webSocket, BusinessEvent: events,
	}
}

func contextOrBackground(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
}
