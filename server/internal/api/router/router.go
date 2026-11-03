package router

import "github.com/gcc798/xianyu.bot/internal/api"

// Router 汇总后端路由模块。
type Router struct {
	JWT              *JWTRouter              // JWT 注册令牌管理路由。
	Auth             *AuthRouter             // Auth 注册认证路由。
	User             *UserRouter             // User 注册用户路由。
	Dictionary       *DictionaryRouter       // Dictionary 注册字典路由。
	DictionaryDetail *DictionaryDetailRouter // DictionaryDetail 注册字典项路由。
	LoginLog         *LoginLogRouter         // LoginLog 注册登录审计路由。
	Status           *StatusRouter           // Status 注册状态路由。
	Account          *AccountRouter          // Account 注册闲鱼账号路由。
	Product          *ProductRouter          // Product 注册商品路由。
	ReplyRule        *ReplyRuleRouter        // ReplyRule 注册回复规则路由。
	DeliveryRule     *DeliveryRuleRouter     // DeliveryRule 注册发货规则路由。
	Message          *MessageRouter          // Message 注册消息路由。
	Order            *OrderRouter            // Order 注册订单路由。
	QRLogin          *QRLoginRouter          // QRLogin 注册扫码登录路由。
	WebSocket        *WebSocketRouter        // WebSocket 注册连接路由。
	BusinessEvent    *BusinessEventRouter    // BusinessEvent 注册业务事件路由。
	TelegramProxy    *TelegramProxyRouter    // TelegramProxy 注册 Telegram 代理路由。
	AIConfig         *AIConfigRouter         // AIConfig 注册 Agent 配置路由。
	Resource         *ResourceRouter         // Resource 注册资源路由。
	AgentSession     *AgentSessionRouter     // AgentSession 注册 Agent 会话路由。
}

// New 创建路由模块集合。
func New(handlers *api.API) *Router {
	return &Router{
		JWT: NewJWTRouter(handlers), Auth: NewAuthRouter(handlers), User: NewUserRouter(handlers),
		Dictionary: NewDictionaryRouter(handlers), DictionaryDetail: NewDictionaryDetailRouter(handlers), LoginLog: NewLoginLogRouter(handlers),
		Status: NewStatusRouter(handlers), Account: NewAccountRouter(handlers), Product: NewProductRouter(handlers),
		ReplyRule: NewReplyRuleRouter(handlers), DeliveryRule: NewDeliveryRuleRouter(handlers), Message: NewMessageRouter(handlers),
		Order: NewOrderRouter(handlers), QRLogin: NewQRLoginRouter(handlers), WebSocket: NewWebSocketRouter(handlers),
		BusinessEvent: NewBusinessEventRouter(handlers), TelegramProxy: NewTelegramProxyRouter(handlers), AIConfig: NewAIConfigRouter(handlers),
		Resource: NewResourceRouter(handlers), AgentSession: NewAgentSessionRouter(handlers),
	}
}
