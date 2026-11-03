package api

import (
	"github.com/gcc798/xianyu.bot/internal/agent"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gcc798/xianyu.bot/internal/service"
)

type UserAPI struct {
	service *service.Service
	jwt     *security.JWT
}

func NewUserAPI(services *service.Service, jwt *security.JWT) *UserAPI {
	return &UserAPI{service: services, jwt: jwt}
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

func New(
	services *service.Service,
	jwt *security.JWT,
	aiConfig *agent.AIConfig,
	resources *agent.ResourceCatalog,
	sessions *agent.Session,
	runner *agent.Runner,
	events *agent.EventHub,
) *API {
	return &API{
		JWT: NewJWTAPI(services), User: NewUserAPI(services, jwt),
		Dictionary: NewDictionaryAPI(services), DictionaryDetail: NewDictionaryDetailAPI(services), LoginLog: NewLoginLogAPI(services),
		Status: NewStatusAPI(services), Account: NewAccountAPI(services), Product: NewProductAPI(services),
		ReplyRule: NewReplyRuleAPI(services), DeliveryRule: NewDeliveryRuleAPI(services), Message: NewMessageAPI(services),
		Order: NewOrderAPI(services), QRLogin: NewQRLoginAPI(services), WebSocket: NewWebSocketAPI(services),
		BusinessEvent: NewBusinessEventAPI(services), TelegramProxy: NewTelegramProxyAPI(), AIConfig: NewAIConfigAPI(aiConfig),
		Resource: NewResourceAPI(resources), AgentSession: NewAgentSessionAPI(sessions, runner, events),
	}
}
