package api

import (
	"github.com/gcc798/xianyu.bot/internal/agent"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gcc798/xianyu.bot/internal/service"
)

// UserAPI 处理用户 HTTP 请求。
type UserAPI struct {
	svc *service.Service // svc 提供用户业务能力。
	jwt *security.JWT    // jwt 签发登录令牌。
}

// NewUserAPI 创建 UserAPI。
func NewUserAPI(svc *service.Service, jwt *security.JWT) *UserAPI {
	return &UserAPI{svc: svc, jwt: jwt}
}

// API 汇总后端 HTTP 处理器。
type API struct {
	User             *UserAPI              // User 处理用户请求。
	Dictionary       *DictionaryAPI        // Dictionary 处理字典请求。
	DictionaryDetail *DictionaryDetailAPI  // DictionaryDetail 处理字典项请求。
	LoginLog         *LoginLogAPI          // LoginLog 处理登录审计请求。
	Status           *StatusAPI            // Status 处理状态请求。
	Account          *AccountAPI           // Account 处理闲鱼账号请求。
	Product          *ProductAPI           // Product 处理商品请求。
	ReplyRule        *ReplyRuleAPI         // ReplyRule 处理回复规则请求。
	DeliveryRule     *DeliveryRuleAPI      // DeliveryRule 处理发货规则请求。
	Message          *MessageAPI           // Message 处理消息请求。
	Order            *OrderAPI             // Order 处理订单请求。
	QRLogin          *QRLoginAPI           // QRLogin 处理扫码登录请求。
	WebSocket        *WebSocketAPI         // WebSocket 处理连接请求。
	BusinessEvent    *BusinessEventAPI     // BusinessEvent 处理业务事件请求。
	TelegramProxy    *TelegramProxyAPI     // TelegramProxy 代理 Telegram 请求。
	AIConfig         *AIConfigAPI          // AIConfig 处理 Agent 配置请求。
	TelegramAPI      *TelegramAPIConfigAPI // TelegramAPI 管理 Telegram 应用凭据。
	Resource         *ResourceAPI          // Resource 处理资源请求。
	AgentSession     *AgentSessionAPI      // AgentSession 处理 Agent 会话请求。
}

// New 创建 HTTP 处理器集合。
func New(
	svc *service.Service,
	jwt *security.JWT,
	aiConfig *agent.AIConfig,
	resources *agent.ResourceCatalog,
	sessions *agent.Session,
	runner *agent.Runner,
	events *agent.EventHub,
) *API {
	return &API{
		User:             NewUserAPI(svc, jwt),
		Dictionary:       NewDictionaryAPI(svc),
		DictionaryDetail: NewDictionaryDetailAPI(svc),
		LoginLog:         NewLoginLogAPI(svc),
		Status:           NewStatusAPI(svc),
		Account:          NewAccountAPI(svc),
		Product:          NewProductAPI(svc),
		ReplyRule:        NewReplyRuleAPI(svc),
		DeliveryRule:     NewDeliveryRuleAPI(svc),
		Message:          NewMessageAPI(svc),
		Order:            NewOrderAPI(svc),
		QRLogin:          NewQRLoginAPI(svc),
		WebSocket:        NewWebSocketAPI(svc),
		BusinessEvent:    NewBusinessEventAPI(svc),
		TelegramProxy:    NewTelegramProxyAPI(),
		AIConfig:         NewAIConfigAPI(aiConfig),
		TelegramAPI:      NewTelegramAPIConfigAPI(svc.TelegramAPI),
		Resource:         NewResourceAPI(resources),
		AgentSession:     NewAgentSessionAPI(sessions, runner, events),
	}
}
