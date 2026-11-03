package router

import "github.com/gcc798/xianyu.bot/internal/api"

type Router struct {
	JWT              *JWTRouter
	Auth             *AuthRouter
	User             *UserRouter
	Dictionary       *DictionaryRouter
	DictionaryDetail *DictionaryDetailRouter
	LoginLog         *LoginLogRouter
	Status           *StatusRouter
	Account          *AccountRouter
	Product          *ProductRouter
	ReplyRule        *ReplyRuleRouter
	DeliveryRule     *DeliveryRuleRouter
	Message          *MessageRouter
	Order            *OrderRouter
	QRLogin          *QRLoginRouter
	WebSocket        *WebSocketRouter
	BusinessEvent    *BusinessEventRouter
	TelegramProxy    *TelegramProxyRouter
	AIConfig         *AIConfigRouter
	Resource         *ResourceRouter
	AgentSession     *AgentSessionRouter
}

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
