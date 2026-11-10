package service

import (
	"context"

	"gorm.io/gorm"
)

// Service 汇总 HTTP 层直接使用的系统与闲鱼业务服务。
type Service struct {
	RevokedTokens    *RevokedTokenService     // RevokedTokens 管理登录令牌撤销。
	User             *UserService             // User 管理系统用户。
	Dictionary       *DictionaryService       // Dictionary 管理系统字典。
	DictionaryDetail *DictionaryDetailService // DictionaryDetail 管理字典条目。
	LoginLog         *LoginLogService         // LoginLog 管理登录审计。
	Status           *Status                  // Status 提供健康状态与仪表盘统计。
	Account          *Account                 // Account 管理闲鱼账号与凭据。
	Product          *Product                 // Product 管理闲鱼商品。
	ReplyRule        *ReplyRule               // ReplyRule 管理自动回复规则。
	DeliveryRule     *DeliveryRule            // DeliveryRule 管理自动发货规则。
	Delivery         *Delivery                // Delivery 调度自动发货。
	Relist           *Relist                  // Relist 调度商品自动重发。
	Order            *Order                   // Order 管理闲鱼订单。
	Message          *Message                 // Message 管理闲鱼消息与自动回复。
	QRLogin          *QRLogin                 // QRLogin 管理二维码登录会话。
	WebSocket        *WebSocketService        // WebSocket 管理闲鱼消息连接。
	BusinessEvent    *BusinessEvent           // BusinessEvent 记录业务审计事件。
	TelegramAPI      *TelegramAPIConfig       // TelegramAPI 管理 Telegram 应用凭据。
}

// New 组装共享同一数据库和根上下文的全部业务服务。
func New(ctx context.Context, db *gorm.DB, sendNotificationEmail func(string, string) error) *Service {
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
		TelegramAPI: NewTelegramAPIConfig(db),
	}
}

// contextOrBackground 保留有效根上下文，nil 时回退到后台上下文。
func contextOrBackground(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
}
