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

func NewJWTRouter(handlers *api.API) *JWTRouter   { return &JWTRouter{handlers: handlers} }
func NewAuthRouter(handlers *api.API) *AuthRouter { return &AuthRouter{handlers: handlers} }
func NewUserRouter(handlers *api.API) *UserRouter { return &UserRouter{handlers: handlers} }
func NewDictionaryRouter(handlers *api.API) *DictionaryRouter {
	return &DictionaryRouter{handlers: handlers}
}
func NewDictionaryDetailRouter(handlers *api.API) *DictionaryDetailRouter {
	return &DictionaryDetailRouter{handlers: handlers}
}
func NewLoginLogRouter(handlers *api.API) *LoginLogRouter {
	return &LoginLogRouter{handlers: handlers}
}
func NewStatusRouter(handlers *api.API) *StatusRouter { return &StatusRouter{handlers: handlers} }
func NewAccountRouter(handlers *api.API) *AccountRouter {
	return &AccountRouter{handlers: handlers}
}
func NewProductRouter(handlers *api.API) *ProductRouter {
	return &ProductRouter{handlers: handlers}
}
func NewReplyRuleRouter(handlers *api.API) *ReplyRuleRouter {
	return &ReplyRuleRouter{handlers: handlers}
}
func NewDeliveryRuleRouter(handlers *api.API) *DeliveryRuleRouter {
	return &DeliveryRuleRouter{handlers: handlers}
}
func NewMessageRouter(handlers *api.API) *MessageRouter { return &MessageRouter{handlers: handlers} }
func NewOrderRouter(handlers *api.API) *OrderRouter     { return &OrderRouter{handlers: handlers} }
func NewQRLoginRouter(handlers *api.API) *QRLoginRouter {
	return &QRLoginRouter{handlers: handlers}
}
func NewWebSocketRouter(handlers *api.API) *WebSocketRouter {
	return &WebSocketRouter{handlers: handlers}
}
func NewBusinessEventRouter(handlers *api.API) *BusinessEventRouter {
	return &BusinessEventRouter{handlers: handlers}
}
func NewTelegramProxyRouter(handlers *api.API) *TelegramProxyRouter {
	return &TelegramProxyRouter{handlers: handlers}
}
func NewAIConfigRouter(handlers *api.API) *AIConfigRouter {
	return &AIConfigRouter{handlers: handlers}
}
func NewResourceRouter(handlers *api.API) *ResourceRouter {
	return &ResourceRouter{handlers: handlers}
}
func NewAgentSessionRouter(handlers *api.API) *AgentSessionRouter {
	return &AgentSessionRouter{handlers: handlers}
}
