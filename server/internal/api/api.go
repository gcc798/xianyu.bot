package api

import (
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gcc798/xianyu.bot/internal/service"
)

type UserAPI struct {
	service *service.Service
	jwt     *security.JWT
}

type API struct {
	JWT              *JWTAPI
	User             *UserAPI
	Dictionary       *DictionaryAPI
	DictionaryDetail *DictionaryDetailAPI
	LoginLog         *LoginLogAPI
	Status           *StatusAPI
	Account          *AccountAPI
	Product          *ProductAPI
	ReplyRule        *ReplyRuleAPI
	DeliveryRule     *DeliveryRuleAPI
	Message          *MessageAPI
	Order            *OrderAPI
	QRLogin          *QRLoginAPI
	WebSocket        *WebSocketAPI
	BusinessEvent    *BusinessEventAPI
	TelegramProxy    *TelegramProxyAPI
	AIConfig         *AIConfigAPI
	Resource         *ResourceAPI
	AgentSession     *AgentSessionAPI
}

func New(services *service.Service, jwt *security.JWT) *API {
	return &API{
		JWT: NewJWTAPI(services), User: NewUserAPI(services, jwt),
		Dictionary: NewDictionaryAPI(services), DictionaryDetail: NewDictionaryDetailAPI(services), LoginLog: NewLoginLogAPI(services),
		Status: NewStatusAPI(services), Account: NewAccountAPI(services), Product: NewProductAPI(services),
		ReplyRule: NewReplyRuleAPI(services), DeliveryRule: NewDeliveryRuleAPI(services), Message: NewMessageAPI(services),
		Order: NewOrderAPI(services), QRLogin: NewQRLoginAPI(services), WebSocket: NewWebSocketAPI(services),
		BusinessEvent: NewBusinessEventAPI(services), TelegramProxy: NewTelegramProxyAPI(), AIConfig: NewAIConfigAPI(services),
		Resource: NewResourceAPI(services), AgentSession: NewAgentSessionAPI(services),
	}
}

func NewJWTAPI(services *service.Service) *JWTAPI { return &JWTAPI{service: services} }
func NewUserAPI(services *service.Service, jwt *security.JWT) *UserAPI {
	return &UserAPI{service: services, jwt: jwt}
}
func NewDictionaryAPI(services *service.Service) *DictionaryAPI {
	return &DictionaryAPI{service: services}
}
func NewDictionaryDetailAPI(services *service.Service) *DictionaryDetailAPI {
	return &DictionaryDetailAPI{service: services}
}
func NewLoginLogAPI(services *service.Service) *LoginLogAPI {
	return &LoginLogAPI{service: services}
}
func NewStatusAPI(services *service.Service) *StatusAPI { return &StatusAPI{service: services} }
func NewAccountAPI(services *service.Service) *AccountAPI {
	return &AccountAPI{service: services}
}
func NewProductAPI(services *service.Service) *ProductAPI {
	return &ProductAPI{service: services}
}
func NewReplyRuleAPI(services *service.Service) *ReplyRuleAPI {
	return &ReplyRuleAPI{service: services}
}
func NewDeliveryRuleAPI(services *service.Service) *DeliveryRuleAPI {
	return &DeliveryRuleAPI{service: services}
}
func NewMessageAPI(services *service.Service) *MessageAPI { return &MessageAPI{service: services} }
func NewOrderAPI(services *service.Service) *OrderAPI     { return &OrderAPI{service: services} }
func NewQRLoginAPI(services *service.Service) *QRLoginAPI {
	return &QRLoginAPI{service: services}
}
func NewWebSocketAPI(services *service.Service) *WebSocketAPI {
	return &WebSocketAPI{service: services}
}
func NewBusinessEventAPI(services *service.Service) *BusinessEventAPI {
	return &BusinessEventAPI{service: services}
}
func NewTelegramProxyAPI() *TelegramProxyAPI { return new(TelegramProxyAPI) }
func NewAIConfigAPI(services *service.Service) *AIConfigAPI {
	return &AIConfigAPI{service: services}
}
func NewResourceAPI(services *service.Service) *ResourceAPI {
	return &ResourceAPI{service: services}
}
func NewAgentSessionAPI(services *service.Service) *AgentSessionAPI {
	return &AgentSessionAPI{service: services}
}
