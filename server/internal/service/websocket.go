package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gcc798/xianyu.bot/internal/xianyu/connection"
	"github.com/gcc798/xianyu.bot/internal/xianyu/protocol"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
	"gorm.io/gorm"
)

const (
	webSocketEndpoint       = "wss://wss-goofish.dingtalk.com/"
	webBrowserUA            = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"
	webSocketRegistrationUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
)

type webSocketTokenGetter interface {
	Get(context.Context, uint, uint) (string, string, error)
}

type WebSocketService struct {
	root        context.Context
	db          *gorm.DB
	events      *BusinessEvent
	manager     *connection.Manager
	accounts    *Account
	messages    *Message
	keepaliveMu sync.Mutex
	keepalives  map[uint]context.CancelFunc
	alertMu     sync.Mutex
	// shortcut: active alert windows reset on process restart; persist them if cross-restart continuity becomes required.
	alertWindows          map[uint]*connectionAlertWindow
	alertWindowDuration   time.Duration
	sendNotificationEmail func(string, string, string) error
	delivery              interface {
		StartAccount(uint)
		StopAccount(uint)
	}
}

type xyDialer struct {
	db       *gorm.DB
	token    webSocketTokenGetter
	endpoint string
}

type xyConnection struct {
	accountID uint
	conn      *websocket.Conn
	writeMu   sync.Mutex
	pendingMu sync.Mutex
	pending   map[string]chan webSocketResponse
	closeOnce sync.Once
}

type webSocketResponse struct {
	code int
	raw  []byte
}

func NewWebSocketService(root context.Context, db *gorm.DB, events *BusinessEvent, messages *Message, sendNotificationEmail func(string, string, string) error) *WebSocketService {
	token := newWebSocketToken(db, events)
	dialer := &xyDialer{db: db, token: &token, endpoint: webSocketEndpoint}
	service := &WebSocketService{
		root:                  contextOrBackground(root),
		db:                    db,
		events:                events,
		messages:              messages,
		alertWindows:          make(map[uint]*connectionAlertWindow),
		alertWindowDuration:   time.Hour,
		sendNotificationEmail: sendNotificationEmail,
	}
	config := connection.DefaultConfig()
	config.StateChanged = service.connectionStateChanged
	manager, err := connection.NewManager(dialer, service.handleRuntimeMessage, config)
	if err != nil {
		panic(err)
	}
	service.manager = manager
	return service
}

func (s *WebSocketService) handleRuntimeMessage(ctx context.Context, accountID uint, raw []byte) error {
	if isSyncChangeSignal(raw) {
		s.syncRecentMessages(ctx, accountID)
		return nil
	}
	return s.messages.Handle(ctx, accountID, raw)
}

func isSyncChangeSignal(raw []byte) bool {
	var envelope struct {
		LWP  string `json:"lwp"`
		Body struct {
			SyncExtraType struct {
				Type int `json:"type"`
			} `json:"syncExtraType"`
		} `json:"body"`
	}
	return json.Unmarshal(raw, &envelope) == nil && envelope.LWP == protocol.SyncPush && envelope.Body.SyncExtraType.Type == 2
}

func (s *WebSocketService) Start(ctx context.Context, ownerID, accountID uint) error {
	account, err := ownedAccount(ctx, s.db, ownerID, accountID)
	if err != nil {
		return err
	}
	if !account.Enabled {
		return errors.New("闲鱼账号已停用")
	}
	if account.Status != "ready" {
		return errors.New("闲鱼账号登录状态不可用")
	}
	if err := s.manager.Start(contextOrBackground(s.root), accountID); err != nil {
		return err
	}
	if s.delivery != nil {
		s.delivery.StartAccount(accountID)
	}
	if s.messages != nil {
		s.messages.StartAccount(accountID)
	}
	s.startKeepalive(ownerID, accountID)
	if snapshot, running := s.manager.Snapshot(accountID); !running && snapshot.State == connection.StateFailed {
		s.stopAccountWorkers(accountID)
	}
	return nil
}

func (s *WebSocketService) StartEnabledAccounts(ctx context.Context) {
	var accounts []model.Account
	if err := s.db.WithContext(ctx).Where("enabled = ? AND status = ?", true, "ready").Find(&accounts).Error; err != nil {
		zap.L().Error("加载需自动连接的闲鱼账号失败", zap.Error(err))
		return
	}
	for _, account := range accounts {
		if err := s.Start(ctx, account.OwnerID, account.ID); err != nil {
			zap.L().Error("自动连接闲鱼账号失败", zap.Uint("accountID", account.ID), zap.Error(err))
		}
	}
}

func (s *WebSocketService) Stop(ctx context.Context, ownerID, accountID uint) error {
	if _, err := ownedAccount(ctx, s.db, ownerID, accountID); err != nil {
		return err
	}
	if !s.stopRuntime(accountID) {
		return errors.New("账号连接未运行")
	}
	return nil
}

func (s *WebSocketService) stopRuntime(accountID uint) bool {
	s.stopAccountWorkers(accountID)
	s.cancelConnectionAlertWindow(accountID)
	return s.manager.Stop(accountID)
}

func (s *WebSocketService) stopAccountWorkers(accountID uint) {
	if s.delivery != nil {
		s.delivery.StopAccount(accountID)
	}
	if s.messages != nil {
		s.messages.StopAccount(accountID)
	}
	s.stopKeepalive(accountID)
}

func (s *WebSocketService) startKeepalive(ownerID, accountID uint) {
	if s.accounts == nil {
		return
	}
	s.keepaliveMu.Lock()
	if s.keepalives == nil {
		s.keepalives = make(map[uint]context.CancelFunc)
	}
	if s.keepalives[accountID] != nil {
		s.keepaliveMu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(contextOrBackground(s.root))
	s.keepalives[accountID] = cancel
	s.keepaliveMu.Unlock()
	go func() {
		loginTicker := time.NewTicker(30 * time.Minute)
		historyTicker := time.NewTicker(time.Minute)
		initialHistory := time.NewTimer(5 * time.Second)
		defer loginTicker.Stop()
		defer historyTicker.Stop()
		defer initialHistory.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-initialHistory.C:
				s.syncRecentMessages(ctx, accountID)
			case <-historyTicker.C:
				s.syncRecentMessages(ctx, accountID)
			case <-loginTicker.C:
				checkCtx, checkCancel := context.WithTimeout(ctx, 30*time.Second)
				result, err := s.accounts.CheckLogin(checkCtx, ownerID, accountID)
				checkCancel()
				if err == nil && result.Status == "ready" {
					continue
				}
				if s.delivery != nil {
					s.delivery.StopAccount(accountID)
				}
				s.manager.Stop(accountID)
				s.stopKeepalive(accountID)
				return
			}
		}
	}()
}

func (s *WebSocketService) syncRecentMessages(ctx context.Context, accountID uint) {
	if s.messages == nil {
		return
	}
	requestCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	listPayload, err := buildIMRequest("/r/Conversation/listNewestPagination", []any{int64(9007199254740991), 5})
	if err != nil {
		zap.L().Warn("构造闲鱼最近会话同步请求失败", zap.Uint("accountID", accountID), zap.Error(err))
		return
	}
	response, err := s.manager.Request(requestCtx, accountID, listPayload)
	if err != nil {
		if requestCtx.Err() == nil {
			zap.L().Warn("请求闲鱼最近会话失败", zap.Uint("accountID", accountID), zap.Error(err))
		}
		return
	}
	conversations, err := protocol.ParseConversationResponse(response)
	if err != nil {
		zap.L().Warn("闲鱼最近会话同步失败", zap.Uint("accountID", accountID), zap.Error(err))
		return
	}
	for _, sessionID := range conversations {
		if requestCtx.Err() != nil {
			return
		}
		messagePayload, err := buildIMRequest("/r/MessageManager/listUserMessages", []any{ensureGoofishID(sessionID), false, int64(9007199254740991), 20, false})
		if err != nil {
			zap.L().Warn("构造闲鱼历史消息同步请求失败", zap.Uint("accountID", accountID), zap.String("sessionID", sessionID), zap.Error(err))
			return
		}
		history, err := s.manager.Request(requestCtx, accountID, messagePayload)
		if err != nil {
			if requestCtx.Err() == nil {
				zap.L().Warn("请求闲鱼历史消息失败", zap.Uint("accountID", accountID), zap.String("sessionID", sessionID), zap.Error(err))
			}
			return
		}
		if err := s.messages.HandleHistory(requestCtx, accountID, sessionID, history); err != nil {
			zap.L().Warn("闲鱼会话消息补偿同步失败", zap.Uint("accountID", accountID), zap.Error(err))
			return
		}
	}
}

func buildIMRequest(lwp string, body []any) ([]byte, error) {
	return json.Marshal(map[string]any{
		"lwp": lwp, "headers": map[string]string{"mid": webSocketMID()}, "body": body,
	})
}

func ensureGoofishID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasSuffix(value, "@goofish") {
		return value
	}
	return value + "@goofish"
}

func (s *WebSocketService) stopKeepalive(accountID uint) {
	s.keepaliveMu.Lock()
	cancel := s.keepalives[accountID]
	delete(s.keepalives, accountID)
	s.keepaliveMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (s *WebSocketService) Status(ctx context.Context, ownerID, accountID uint) (xyResponse.WebSocketStatus, error) {
	account, err := ownedAccount(ctx, s.db, ownerID, accountID)
	if err != nil {
		return xyResponse.WebSocketStatus{}, err
	}
	snapshot, running := s.manager.Snapshot(accountID)
	return xyResponse.WebSocketStatus{
		AccountID: accountID, State: string(snapshot.State), QueueDepth: snapshot.QueueDepth,
		Reconnects: snapshot.Reconnects, Running: running, LastError: snapshot.LastError,
		AccountStatus: account.Status,
	}, nil
}

func (s *WebSocketService) SendText(ctx context.Context, ownerID, accountID uint, sessionID, toUserID, text string) error {
	account, err := ownedAccount(ctx, s.db, ownerID, accountID)
	if err != nil {
		return err
	}
	text = strings.TrimSpace(text)
	if sessionID == "" || toUserID == "" || text == "" || len([]rune(text)) > 4000 {
		return errors.New("会话、接收方或消息内容无效")
	}
	if account.ExternalUserID == "" {
		return errors.New("闲鱼账号缺少用户标识，请重新检查登录态")
	}
	payload, err := buildTextMessage(sessionID, toUserID, account.ExternalUserID, text)
	if err != nil {
		return err
	}
	sendCtx, cancel := context.WithTimeoutCause(ctx, 10*time.Second, errors.New("闲鱼消息发送结果未知"))
	defer cancel()
	err = s.manager.Send(sendCtx, accountID, payload)
	fields := []zap.Field{zap.Uint("ownerID", ownerID), zap.Uint("accountID", accountID), zap.String("sessionID", sessionID), zap.String("toUserID", toUserID), zap.Int("textLength", len([]rune(text)))}
	if err != nil {
		zap.L().Warn("发送闲鱼消息失败", append(fields, zap.Error(err))...)
	} else {
		zap.L().Info("发送闲鱼消息成功", fields...)
	}
	return err
}

func buildTextMessage(sessionID, toUserID, fromUserID, text string) ([]byte, error) {
	content, err := json.Marshal(map[string]any{"contentType": 1, "text": map[string]string{"text": text}})
	if err != nil {
		return nil, err
	}
	clean := func(value string) string { return strings.TrimSuffix(strings.TrimSpace(value), "@goofish") }
	message := map[string]any{
		"lwp":     "/r/MessageSend/sendByReceiverScope",
		"headers": map[string]string{"mid": webSocketMID()},
		"body": []any{
			map[string]any{
				"uuid": webSocketMID(), "cid": clean(sessionID) + "@goofish", "conversationType": 1,
				"content": map[string]any{"contentType": 101, "custom": map[string]any{
					"type": 1, "data": base64.StdEncoding.EncodeToString(content),
				}},
				"redPointPolicy": 0, "extension": map[string]string{"extJson": "{}"},
				"ctx":   map[string]string{"appVersion": "1.0", "platform": "web"},
				"mtags": map[string]any{}, "msgReadStatusSetting": 1,
			},
			map[string]any{"actualReceivers": []string{clean(toUserID) + "@goofish", clean(fromUserID) + "@goofish"}},
		},
	}
	return json.Marshal(message)
}

func ownedAccount(ctx context.Context, db *gorm.DB, ownerID, accountID uint) (model.Account, error) {
	if ownerID == 0 || accountID == 0 {
		return model.Account{}, errors.New("账号 ID 不能为空")
	}
	var account model.Account
	if err := db.WithContext(ctx).Where("id = ? AND owner_id = ?", accountID, ownerID).First(&account).Error; err != nil {
		return model.Account{}, err
	}
	return account, nil
}
