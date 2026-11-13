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
	// webSocketEndpoint 是闲鱼消息 WebSocket 入口。
	webSocketEndpoint = "wss://wss-goofish.dingtalk.com/"
	// webBrowserUA 是 HTTP 与握手请求使用的浏览器标识。
	webBrowserUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"
	// webSocketRegistrationUA 是闲鱼消息协议注册字段要求的客户端标识。
	webSocketRegistrationUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
)

// webSocketTokenGetter 描述连接拨号前获取令牌和设备标识的能力。
type webSocketTokenGetter interface {
	// Get 为 ownerID 名下账号返回有效消息令牌和设备标识。
	Get(context.Context, uint, uint) (string, string, error)
}

// WebSocketService 编排闲鱼消息连接、补偿同步和账号后台任务。
type WebSocketService struct {
	root        context.Context             // root 控制连接相关后台任务生命周期。
	db          *gorm.DB                    // db 读取账号并汇总连接状态。
	events      *BusinessEvent              // events 记录连接业务事件。
	manager     *connection.Manager         // manager 管理每个账号的连接运行时。
	accounts    *Account                    // accounts 周期检查登录状态。
	messages    *Message                    // messages 解析并持久化实时与历史消息。
	keepaliveMu sync.Mutex                  // keepaliveMu 保护 keepalives。
	keepalives  map[uint]context.CancelFunc // keepalives 保存账号保活任务取消函数。
	alertMu     sync.Mutex                  // alertMu 保护 alertWindows。
	// shortcut: active alert windows reset on process restart; persist them if cross-restart continuity becomes required.
	alertWindows          map[uint]*ConnectionAlertWindow // alertWindows 聚合账号连接抖动。
	alertWindowDuration   time.Duration                   // alertWindowDuration 是告警聚合周期。
	sendNotificationEmail func(string, string) error      // sendNotificationEmail 发送故障与恢复通知。
	delivery              interface {
		// StartAccount 启动账号发货后台任务。
		StartAccount(uint)
		// StopAccount 停止账号发货后台任务。
		StopAccount(uint)
	} // delivery 管理与连接同生命周期的发货执行器。
}

// NewWebSocketService 创建消息连接服务并组装令牌、拨号和运行时管理器。
func NewWebSocketService(root context.Context, db *gorm.DB, events *BusinessEvent, messages *Message, sendNotificationEmail func(string, string) error) *WebSocketService {
	token := NewWebSocketToken(db, events)
	dialer := &XYDialer{db: db, token: &token, endpoint: webSocketEndpoint}
	svc := &WebSocketService{
		root:                  contextOrBackground(root),
		db:                    db,
		events:                events,
		messages:              messages,
		alertWindows:          make(map[uint]*ConnectionAlertWindow),
		alertWindowDuration:   time.Hour,
		sendNotificationEmail: sendNotificationEmail,
	}
	config := connection.DefaultConfig()
	config.StateChanged = svc.connectionStateChanged
	manager, err := connection.NewManager(dialer, svc.handleRuntimeMessage, config)
	if err != nil {
		panic(err)
	}
	svc.manager = manager
	return svc
}

// XYDialer 使用账号 Cookie 和消息令牌建立闲鱼连接。
type XYDialer struct {
	db       *gorm.DB             // db 读取账号和加密 Cookie。
	token    webSocketTokenGetter // token 获取有效消息令牌与设备标识。
	endpoint string               // endpoint 是可在测试中替换的连接地址。
}

// XYConnection 封装单个账号的 WebSocket 连接和请求响应关联。
type XYConnection struct {
	accountID uint                              // accountID 是连接所属账号 ID。
	conn      *websocket.Conn                   // conn 是底层 WebSocket 连接。
	writeMu   sync.Mutex                        // writeMu 串行化协议帧写入。
	pendingMu sync.Mutex                        // pendingMu 保护 pending。
	pending   map[string]chan WebSocketResponse // pending 按消息 ID 关联等待中的请求。
	closeOnce sync.Once                         // closeOnce 保证底层连接只关闭一次。
}

// WebSocketResponse 是请求消息对应的协议响应。
type WebSocketResponse struct {
	code int    // code 是闲鱼协议状态码。
	raw  []byte // raw 是完整响应帧。
}

// handleRuntimeMessage 将同步变化信号转为补偿同步，其余帧交给消息服务。
func (s *WebSocketService) handleRuntimeMessage(ctx context.Context, accountID uint, raw []byte) error {
	if isSyncChangeSignal(raw) {
		s.syncRecentMessages(ctx, accountID)
		return nil
	}
	return s.messages.Handle(ctx, accountID, raw)
}

// isSyncChangeSignal 判断协议帧是否要求主动同步最近会话。
func isSyncChangeSignal(raw []byte) bool {
	var envelope struct {
		LWP  string `json:"lwp"` // LWP 是协议方法名。
		Body struct {
			SyncExtraType struct {
				Type int `json:"type"` // Type 为 2 时表示会话列表变化。
			} `json:"syncExtraType"` // SyncExtraType 描述同步变化类型。
		} `json:"body"` // Body 是协议业务负载。
	}
	return json.Unmarshal(raw, &envelope) == nil && envelope.LWP == protocol.SyncPush && envelope.Body.SyncExtraType.Type == 2
}

// Start 校验账号归属与登录态，并启动连接、发货、回复和保活任务。
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

// StartEnabledAccounts 在服务启动时连接全部启用且就绪的账号。
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

// Stop 校验账号归属并停止其连接及关联后台任务。
func (s *WebSocketService) Stop(ctx context.Context, ownerID, accountID uint) error {
	if _, err := ownedAccount(ctx, s.db, ownerID, accountID); err != nil {
		return err
	}
	if !s.stopRuntime(accountID) {
		return errors.New("账号连接未运行")
	}
	return nil
}

// stopRuntime 停止账号后台任务、告警窗口和连接运行时。
func (s *WebSocketService) stopRuntime(accountID uint) bool {
	s.stopAccountWorkers(accountID)
	s.cancelConnectionAlertWindow(accountID)
	return s.manager.Stop(accountID)
}

// stopAccountWorkers 停止与账号连接同生命周期的业务执行器。
func (s *WebSocketService) stopAccountWorkers(accountID uint) {
	if s.delivery != nil {
		s.delivery.StopAccount(accountID)
	}
	if s.messages != nil {
		s.messages.StopAccount(accountID)
	}
	s.stopKeepalive(accountID)
}

// startKeepalive 启动账号历史补偿同步和半小时登录态检查。
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

// syncRecentMessages 拉取最近会话及各会话最新消息，补偿可能遗漏的推送。
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

// buildIMRequest 生成带唯一消息 ID 的闲鱼 IM 请求帧。
func buildIMRequest(lwp string, body []any) ([]byte, error) {
	return json.Marshal(map[string]any{
		"lwp": lwp, "headers": map[string]string{"mid": webSocketMID()}, "body": body,
	})
}

// ensureGoofishID 为非空用户标识补充 @goofish 后缀。
func ensureGoofishID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasSuffix(value, "@goofish") {
		return value
	}
	return value + "@goofish"
}

// stopKeepalive 取消并移除指定账号的保活任务。
func (s *WebSocketService) stopKeepalive(accountID uint) {
	s.keepaliveMu.Lock()
	cancel := s.keepalives[accountID]
	delete(s.keepalives, accountID)
	s.keepaliveMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// Status 校验账号归属并返回连接运行时快照。
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

var errSendUnknown = errors.New("闲鱼消息发送结果未知")

// SendText 校验账号归属和文本边界后向指定闲鱼会话发送消息。
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
	sendCtx, cancel := context.WithTimeoutCause(ctx, 10*time.Second, errSendUnknown)
	defer cancel()
	err = s.manager.Send(sendCtx, accountID, payload)
	if err != nil && !errors.Is(err, connection.ErrNotConnected) {
		err = errors.Join(errSendUnknown, err)
	}
	fields := []zap.Field{zap.Uint("ownerID", ownerID), zap.Uint("accountID", accountID), zap.String("sessionID", sessionID), zap.String("toUserID", toUserID), zap.Int("textLength", len([]rune(text)))}
	if err != nil {
		zap.L().Warn("发送闲鱼消息失败", append(fields, zap.Error(err))...)
	} else {
		zap.L().Info("发送闲鱼消息成功", fields...)
	}
	return err
}

// buildTextMessage 生成闲鱼 IM 文本消息协议帧。
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

// ownedAccount 查询 ownerID 名下指定闲鱼账号。
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
