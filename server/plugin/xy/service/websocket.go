package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/notification/mail"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model"
	xyResponse "github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model/response"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/protocol"
	xyRuntime "github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/runtime"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

const (
	webSocketEndpoint       = "wss://wss-goofish.dingtalk.com/"
	webBrowserUA            = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"
	webSocketRegistrationUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
)

type webSocketTokenGetter interface {
	Get(context.Context, uint, uint) (string, string, error)
}

type webSocketService struct {
	manager     *xyRuntime.Manager
	accounts    *account
	messages    *message
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

type connectionAlertWindow struct {
	startedAt        time.Time
	unavailable      bool
	unavailableCount uint
	alerted          bool
	lastFailure      xyRuntime.Snapshot
	notificationDone chan struct{}
}

type xyDialer struct {
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

func newWebSocketService(messages *message) *webSocketService {
	token := newWebSocketToken()
	dialer := &xyDialer{token: &token, endpoint: webSocketEndpoint}
	service := &webSocketService{
		messages:              messages,
		alertWindows:          make(map[uint]*connectionAlertWindow),
		alertWindowDuration:   time.Hour,
		sendNotificationEmail: mail.Send,
	}
	config := xyRuntime.DefaultConfig()
	config.StateChanged = service.connectionStateChanged
	manager, err := xyRuntime.NewManager(dialer, service.handleRuntimeMessage, config)
	if err != nil {
		panic(err)
	}
	service.manager = manager
	return service
}

func (s *webSocketService) connectionStateChanged(snapshot xyRuntime.Snapshot) {
	if global.GVA_LOG != nil {
		fields := []zap.Field{
			zap.Uint("accountID", snapshot.AccountID), zap.String("state", string(snapshot.State)),
			zap.Uint64("reconnects", snapshot.Reconnects), zap.String("lastError", snapshot.LastError),
		}
		if snapshot.State == xyRuntime.StateConnected || snapshot.State == xyRuntime.StateConnecting {
			global.GVA_LOG.Info("闲鱼消息连接状态变化", fields...)
		} else {
			global.GVA_LOG.Warn("闲鱼消息连接状态变化", fields...)
		}
	}

	s.trackConnectionAlertWindow(snapshot)
	if snapshot.State == xyRuntime.StateFailed {
		s.stopAccountWorkers(snapshot.AccountID)
	}
}

func (s *webSocketService) trackConnectionAlertWindow(snapshot xyRuntime.Snapshot) {
	if snapshot.State != xyRuntime.StateConnected && snapshot.State != xyRuntime.StateReconnecting && snapshot.State != xyRuntime.StateFailed {
		return
	}

	s.alertMu.Lock()
	window := s.alertWindows[snapshot.AccountID]
	if snapshot.State == xyRuntime.StateConnected {
		if window == nil {
			s.alertMu.Unlock()
			return
		}
		window.unavailable = false
		if !window.alerted {
			s.alertMu.Unlock()
			return
		}
		delete(s.alertWindows, snapshot.AccountID)
		notificationDone := window.notificationDone
		s.alertMu.Unlock()
		go func() {
			<-notificationDone
			s.sendConnectionNotification(snapshot, true, window.startedAt, window.unavailableCount)
		}()
		return
	}

	if window == nil {
		window = &connectionAlertWindow{startedAt: time.Now()}
		s.alertWindows[snapshot.AccountID] = window
		duration := s.alertWindowDuration
		if duration <= 0 {
			duration = time.Hour
		}
		time.AfterFunc(duration, func() { s.finishConnectionAlertWindow(snapshot.AccountID, window) })
	}
	if !window.unavailable {
		window.unavailableCount++
	}
	window.unavailable = true
	window.lastFailure = snapshot
	s.alertMu.Unlock()
}

func (s *webSocketService) finishConnectionAlertWindow(accountID uint, target *connectionAlertWindow) {
	s.alertMu.Lock()
	window := s.alertWindows[accountID]
	if window != target || window.alerted {
		s.alertMu.Unlock()
		return
	}
	if !window.unavailable {
		delete(s.alertWindows, accountID)
		s.alertMu.Unlock()
		s.recordSuppressedConnectionAlert(accountID, window)
		return
	}
	window.alerted = true
	window.notificationDone = make(chan struct{})
	snapshot := window.lastFailure
	s.alertMu.Unlock()
	s.sendConnectionNotification(snapshot, false, window.startedAt, window.unavailableCount)
	close(window.notificationDone)
}

func (s *webSocketService) recordSuppressedConnectionAlert(accountID uint, window *connectionAlertWindow) {
	var account model.Account
	if err := global.GVA_DB.Where("id = ?", accountID).First(&account).Error; err != nil {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Error("记录闲鱼消息连接抖动失败", zap.Uint("accountID", accountID), zap.Error(err))
		}
		return
	}
	recordBusinessEvent(context.Background(), model.BusinessEvent{
		OwnerID: account.OwnerID, AccountID: account.ID, Type: "websocket_alert", Status: "success",
		Summary: fmt.Sprintf("一小时连接抖动窗口结束，当前连接正常（不可用 %d 次，未发送告警）", window.unavailableCount),
		Error:   window.lastFailure.LastError,
	})
}

func (s *webSocketService) cancelConnectionAlertWindow(accountID uint) {
	s.alertMu.Lock()
	delete(s.alertWindows, accountID)
	s.alertMu.Unlock()
}

func (s *webSocketService) sendConnectionNotification(snapshot xyRuntime.Snapshot, recovered bool, startedAt time.Time, unavailableCount uint) {
	var account model.Account
	if err := global.GVA_DB.Where("id = ?", snapshot.AccountID).First(&account).Error; err != nil {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Error("读取闲鱼消息告警账号失败", zap.Uint("accountID", snapshot.AccountID), zap.Error(err))
		}
		return
	}
	subject := "[闲鱼助手告警] 消息连接不可用"
	body := fmt.Sprintf("账号：%s（ID: %d）<br>状态：%s<br>原因：%s<br>窗口开始：%s<br>一小时内不可用次数：%d<br>时间：%s",
		html.EscapeString(account.Label), account.ID, html.EscapeString(string(snapshot.State)),
		html.EscapeString(snapshot.LastError), startedAt.Format(time.DateTime), unavailableCount, time.Now().Format(time.DateTime))
	summary := "已发送闲鱼消息连接故障邮件"
	if recovered {
		subject = "[闲鱼助手恢复] 消息连接已恢复"
		body = fmt.Sprintf("账号：%s（ID: %d）<br>状态：已恢复正常<br>故障窗口开始：%s<br>一小时内不可用次数：%d<br>时间：%s",
			html.EscapeString(account.Label), account.ID, startedAt.Format(time.DateTime), unavailableCount, time.Now().Format(time.DateTime))
		summary = "已发送闲鱼消息连接恢复邮件"
	} else if snapshot.ActionURL != "" {
		body += "<br>处理方式：请在管理后台点击“重新扫码验证”，并在闲鱼 App 完成安全验证。"
	}
	sender := s.sendNotificationEmail
	if sender == nil {
		sender = mail.Send
	}
	if err := sender("force_gc@163.com", subject, body); err != nil {
		recordBusinessEvent(context.Background(), model.BusinessEvent{
			OwnerID: account.OwnerID, AccountID: account.ID, Type: "websocket_alert", Status: "failed",
			Summary: summary + "失败", Error: err.Error(),
		})
		if global.GVA_LOG != nil {
			global.GVA_LOG.Error(summary+"失败", zap.Uint("accountID", snapshot.AccountID), zap.Error(err))
		}
		return
	}
	event := model.BusinessEvent{
		OwnerID: account.OwnerID, AccountID: account.ID, Type: "websocket_alert", Status: "success",
		Summary: fmt.Sprintf("%s（一小时内不可用 %d 次）", summary, unavailableCount),
	}
	if !recovered {
		event.Error = snapshot.LastError
	}
	recordBusinessEvent(context.Background(), event)
	if global.GVA_LOG != nil {
		global.GVA_LOG.Info(summary, zap.Uint("accountID", snapshot.AccountID), zap.String("recipient", "force_gc@163.com"))
	}
}

func (s *webSocketService) handleRuntimeMessage(ctx context.Context, accountID uint, raw []byte) error {
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

func (s *webSocketService) Start(ctx context.Context, ownerID, accountID uint) error {
	account, err := ownedAccount(ctx, ownerID, accountID)
	if err != nil {
		return err
	}
	if !account.Enabled {
		return errors.New("闲鱼账号已停用")
	}
	if account.Status != "ready" {
		return errors.New("闲鱼账号登录状态不可用")
	}
	// shortcut: connection runtimes live for the process; wire a plugin shutdown hook when GVA exposes one.
	if err := s.manager.Start(context.Background(), accountID); err != nil {
		return err
	}
	if s.delivery != nil {
		s.delivery.StartAccount(accountID)
	}
	if s.messages != nil {
		s.messages.StartAccount(accountID)
	}
	s.startKeepalive(ownerID, accountID)
	if snapshot, running := s.manager.Snapshot(accountID); !running && snapshot.State == xyRuntime.StateFailed {
		s.stopAccountWorkers(accountID)
	}
	return nil
}

func (s *webSocketService) StartEnabledAccounts(ctx context.Context) {
	var accounts []model.Account
	if err := global.GVA_DB.WithContext(ctx).Where("enabled = ? AND status = ?", true, "ready").Find(&accounts).Error; err != nil {
		global.GVA_LOG.Error("加载需自动连接的闲鱼账号失败", zap.Error(err))
		return
	}
	for _, account := range accounts {
		if err := s.Start(ctx, account.OwnerID, account.ID); err != nil {
			global.GVA_LOG.Error("自动连接闲鱼账号失败", zap.Uint("accountID", account.ID), zap.Error(err))
		}
	}
}

func (s *webSocketService) Stop(ctx context.Context, ownerID, accountID uint) error {
	if _, err := ownedAccount(ctx, ownerID, accountID); err != nil {
		return err
	}
	if !s.stopRuntime(accountID) {
		return errors.New("账号连接未运行")
	}
	return nil
}

func (s *webSocketService) stopRuntime(accountID uint) bool {
	s.stopAccountWorkers(accountID)
	s.cancelConnectionAlertWindow(accountID)
	return s.manager.Stop(accountID)
}

func (s *webSocketService) stopAccountWorkers(accountID uint) {
	if s.delivery != nil {
		s.delivery.StopAccount(accountID)
	}
	if s.messages != nil {
		s.messages.StopAccount(accountID)
	}
	s.stopKeepalive(accountID)
}

func (s *webSocketService) startKeepalive(ownerID, accountID uint) {
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
	ctx, cancel := context.WithCancel(context.Background())
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

func (s *webSocketService) syncRecentMessages(ctx context.Context, accountID uint) {
	if s.messages == nil {
		return
	}
	requestCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	listPayload, err := buildIMRequest("/r/Conversation/listNewestPagination", []any{int64(9007199254740991), 5})
	if err != nil {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Warn("构造闲鱼最近会话同步请求失败", zap.Uint("accountID", accountID), zap.Error(err))
		}
		return
	}
	response, err := s.manager.Request(requestCtx, accountID, listPayload)
	if err != nil {
		if global.GVA_LOG != nil && requestCtx.Err() == nil {
			global.GVA_LOG.Warn("请求闲鱼最近会话失败", zap.Uint("accountID", accountID), zap.Error(err))
		}
		return
	}
	conversations, err := protocol.ParseConversationResponse(response)
	if err != nil {
		global.GVA_LOG.Warn("闲鱼最近会话同步失败", zap.Uint("accountID", accountID), zap.Error(err))
		return
	}
	for _, sessionID := range conversations {
		if requestCtx.Err() != nil {
			return
		}
		messagePayload, err := buildIMRequest("/r/MessageManager/listUserMessages", []any{ensureGoofishID(sessionID), false, int64(9007199254740991), 20, false})
		if err != nil {
			if global.GVA_LOG != nil {
				global.GVA_LOG.Warn("构造闲鱼历史消息同步请求失败", zap.Uint("accountID", accountID), zap.String("sessionID", sessionID), zap.Error(err))
			}
			return
		}
		history, err := s.manager.Request(requestCtx, accountID, messagePayload)
		if err != nil {
			if global.GVA_LOG != nil && requestCtx.Err() == nil {
				global.GVA_LOG.Warn("请求闲鱼历史消息失败", zap.Uint("accountID", accountID), zap.String("sessionID", sessionID), zap.Error(err))
			}
			return
		}
		if err := s.messages.HandleHistory(requestCtx, accountID, sessionID, history); err != nil {
			global.GVA_LOG.Warn("闲鱼会话消息补偿同步失败", zap.Uint("accountID", accountID), zap.Error(err))
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

func (s *webSocketService) stopKeepalive(accountID uint) {
	s.keepaliveMu.Lock()
	cancel := s.keepalives[accountID]
	delete(s.keepalives, accountID)
	s.keepaliveMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (s *webSocketService) Status(ctx context.Context, ownerID, accountID uint) (xyResponse.WebSocketStatus, error) {
	account, err := ownedAccount(ctx, ownerID, accountID)
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

func (s *webSocketService) SendText(ctx context.Context, ownerID, accountID uint, sessionID, toUserID, text string) error {
	account, err := ownedAccount(ctx, ownerID, accountID)
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
	if global.GVA_LOG != nil {
		fields := []zap.Field{zap.Uint("ownerID", ownerID), zap.Uint("accountID", accountID), zap.String("sessionID", sessionID), zap.String("toUserID", toUserID), zap.Int("textLength", len([]rune(text)))}
		if err != nil {
			global.GVA_LOG.Warn("发送闲鱼消息失败", append(fields, zap.Error(err))...)
		} else {
			global.GVA_LOG.Info("发送闲鱼消息成功", fields...)
		}
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

func ownedAccount(ctx context.Context, ownerID, accountID uint) (model.Account, error) {
	if ownerID == 0 || accountID == 0 {
		return model.Account{}, errors.New("账号 ID 不能为空")
	}
	var account model.Account
	if err := global.GVA_DB.WithContext(ctx).Where("id = ? AND owner_id = ?", accountID, ownerID).First(&account).Error; err != nil {
		return model.Account{}, err
	}
	return account, nil
}

func (d *xyDialer) Dial(ctx context.Context, accountID uint) (xyRuntime.Connection, error) {
	var account model.Account
	if err := global.GVA_DB.WithContext(ctx).Where("id = ?", accountID).First(&account).Error; err != nil {
		return nil, err
	}
	token, deviceID, err := d.token.Get(ctx, account.OwnerID, accountID)
	if err != nil {
		return nil, err
	}
	var credential model.Credential
	if err := global.GVA_DB.WithContext(ctx).Where("account_id = ?", accountID).First(&credential).Error; err != nil {
		return nil, err
	}
	cookie, err := decryptCredential(credential.CookieCiphertext, credentialAAD(accountID, "cookie"))
	if err != nil {
		return nil, err
	}
	config, err := websocket.NewConfig(d.endpoint, "https://www.goofish.com")
	if err != nil {
		return nil, err
	}
	config.Header = http.Header{
		"Cookie":          {cookie},
		"User-Agent":      {webBrowserUA},
		"Accept-Language": {"zh-CN,zh;q=0.9"},
		"Cache-Control":   {"no-cache"},
		"Pragma":          {"no-cache"},
	}
	conn, err := config.DialContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("连接闲鱼 WebSocket: %w", err)
	}
	connection := &xyConnection{accountID: accountID, conn: conn, pending: make(map[string]chan webSocketResponse)}
	registerCtx, cancel := context.WithTimeoutCause(ctx, 10*time.Second, errors.New("闲鱼 WebSocket 注册超时"))
	defer cancel()
	if err := connection.register(registerCtx, token, deviceID); err != nil {
		_ = connection.Close()
		return nil, err
	}
	return connection, nil
}

func (c *xyConnection) register(ctx context.Context, token, deviceID string) error {
	registration := map[string]any{
		"lwp": "/reg",
		"headers": map[string]any{
			"cache-header": "app-key token ua wv", "app-key": webSocketAppKey,
			"token": token, "ua": webSocketRegistrationUA, "dt": "j", "wv": "im:3,au:3,sy:6",
			"sync": "0,0;0;0;", "did": deviceID, "mid": webSocketMID(),
		},
	}
	if err := c.writeJSON(ctx, registration); err != nil {
		return fmt.Errorf("发送闲鱼 WebSocket 注册消息: %w", err)
	}
	for {
		message, err := c.receive(ctx)
		if err != nil {
			return fmt.Errorf("等待闲鱼 WebSocket 注册响应: %w", err)
		}
		var envelope struct {
			Code    int            `json:"code"`
			Headers map[string]any `json:"headers"`
		}
		if json.Unmarshal(message, &envelope) != nil {
			continue
		}
		if envelope.Code == http.StatusUnauthorized {
			return errors.New("闲鱼 WebSocket Token 已失效")
		}
		if envelope.Code == http.StatusOK && envelope.Headers["reg-sid"] != nil {
			break
		}
	}
	now := time.Now().UnixMilli()
	return c.writeJSON(ctx, map[string]any{
		"lwp": "/r/SyncStatus/ackDiff", "headers": map[string]any{"mid": webSocketMID()},
		"body": []map[string]any{{
			"pipeline": "sync", "tooLong2Tag": "PNM,1", "channel": "sync", "topic": "sync",
			"highPts": 0, "pts": now * 1000, "seq": 0, "timestamp": now,
		}},
	})
}

func (c *xyConnection) Read(ctx context.Context) ([]byte, error) {
	for {
		message, err := c.receive(ctx)
		if err != nil {
			return nil, err
		}
		var envelope struct {
			LWP     string          `json:"lwp"`
			Code    int             `json:"code"`
			Headers map[string]any  `json:"headers"`
			Body    json.RawMessage `json:"body"`
		}
		if err := json.Unmarshal(message, &envelope); err != nil {
			if global.GVA_LOG != nil {
				global.GVA_LOG.Warn("闲鱼 WebSocket 收到无法解析的协议帧", zap.Uint("accountID", c.accountID), zap.Int("bytes", len(message)), zap.Error(err))
			}
			continue
		}
		var pushBody struct {
			SyncPushPackage json.RawMessage `json:"syncPushPackage"`
		}
		_ = json.Unmarshal(envelope.Body, &pushBody)
		if global.GVA_LOG != nil && envelope.LWP != "" && envelope.LWP != "/!" {
			global.GVA_LOG.Info("闲鱼 WebSocket 收到协议帧",
				zap.Uint("accountID", c.accountID),
				zap.String("lwp", envelope.LWP),
				zap.Int("code", envelope.Code),
				zap.Bool("hasSyncPushPackage", len(pushBody.SyncPushPackage) > 0 && string(pushBody.SyncPushPackage) != "null"),
			)
		}
		if envelope.Code == http.StatusUnauthorized {
			return nil, errors.New("闲鱼 WebSocket Token 已失效")
		}
		if mid, ok := envelope.Headers["mid"].(string); ok && c.hasPending(mid) {
			if err := c.ack(ctx, envelope.Headers); err != nil {
				return nil, err
			}
			code := envelope.Code
			if code == 0 && len(envelope.Body) > 0 && string(envelope.Body) != "null" {
				code = http.StatusOK
			}
			c.resolve(mid, webSocketResponse{code: code, raw: message})
			return nil, nil
		}
		hasSyncPushPackage := len(pushBody.SyncPushPackage) > 0 && string(pushBody.SyncPushPackage) != "null"
		if hasSyncPushPackage {
			if err := c.ack(ctx, envelope.Headers); err != nil {
				return nil, err
			}
			return message, nil
		}
		if envelope.LWP == "/s/para" || envelope.LWP == "/s/sync" || envelope.LWP == "/s/vulcan" {
			if err := c.ack(ctx, envelope.Headers); err != nil {
				return nil, err
			}
		}
		if envelope.LWP == protocol.SyncPush {
			return message, nil
		}
		return nil, nil
	}
}

func (c *xyConnection) hasPending(mid string) bool {
	c.pendingMu.Lock()
	defer c.pendingMu.Unlock()
	return c.pending[mid] != nil
}

func (c *xyConnection) Send(ctx context.Context, payload []byte) error {
	_, err := c.Request(ctx, payload)
	return err
}

func (c *xyConnection) Request(ctx context.Context, payload []byte) ([]byte, error) {
	var envelope struct {
		Headers struct {
			MID string `json:"mid"`
		} `json:"headers"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil || envelope.Headers.MID == "" {
		return nil, errors.New("闲鱼发送消息缺少请求标识")
	}
	result := make(chan webSocketResponse, 1)
	c.pendingMu.Lock()
	if _, exists := c.pending[envelope.Headers.MID]; exists {
		c.pendingMu.Unlock()
		return nil, errors.New("闲鱼发送消息请求标识重复")
	}
	c.pending[envelope.Headers.MID] = result
	c.pendingMu.Unlock()
	defer func() {
		c.pendingMu.Lock()
		delete(c.pending, envelope.Headers.MID)
		c.pendingMu.Unlock()
	}()
	if err := c.writeText(ctx, payload); err != nil {
		return nil, err
	}
	select {
	case response := <-result:
		if response.code != http.StatusOK {
			return nil, fmt.Errorf("闲鱼请求返回状态 %d", response.code)
		}
		return response.raw, nil
	case <-ctx.Done():
		return nil, context.Cause(ctx)
	}
}

func (c *xyConnection) resolve(mid string, response webSocketResponse) {
	c.pendingMu.Lock()
	result := c.pending[mid]
	c.pendingMu.Unlock()
	if result != nil {
		select {
		case result <- response:
		default:
		}
	}
}

func (c *xyConnection) Ping(ctx context.Context) error {
	return c.writeJSON(ctx, map[string]any{"lwp": "/!", "headers": map[string]any{"mid": webSocketMID()}})
}

func (c *xyConnection) Close() error {
	var err error
	c.closeOnce.Do(func() { err = c.conn.Close() })
	return err
}

func (c *xyConnection) ack(ctx context.Context, headers map[string]any) error {
	if len(headers) == 0 {
		return nil
	}
	ackHeaders := make(map[string]any)
	for _, key := range []string{"mid", "sid", "app-key", "ua", "dt"} {
		if value, ok := headers[key]; ok {
			ackHeaders[key] = value
		}
	}
	if ackHeaders["mid"] == nil {
		ackHeaders["mid"] = webSocketMID()
	}
	return c.writeJSON(ctx, map[string]any{"code": http.StatusOK, "headers": ackHeaders})
}

func (c *xyConnection) receive(ctx context.Context) ([]byte, error) {
	stop := context.AfterFunc(ctx, func() { _ = c.conn.SetReadDeadline(time.Now()) })
	defer stop()
	if deadline, ok := ctx.Deadline(); ok {
		_ = c.conn.SetReadDeadline(deadline)
	}
	var message []byte
	err := websocket.Message.Receive(c.conn, &message)
	_ = c.conn.SetReadDeadline(time.Time{})
	if err != nil && context.Cause(ctx) != nil {
		return nil, context.Cause(ctx)
	}
	return message, err
}

func (c *xyConnection) writeJSON(ctx context.Context, payload any) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	stop := context.AfterFunc(ctx, func() { _ = c.conn.SetWriteDeadline(time.Now()) })
	defer stop()
	if deadline, ok := ctx.Deadline(); ok {
		_ = c.conn.SetWriteDeadline(deadline)
	}
	err := websocket.JSON.Send(c.conn, payload)
	_ = c.conn.SetWriteDeadline(time.Time{})
	if err != nil && context.Cause(ctx) != nil {
		return context.Cause(ctx)
	}
	return err
}

func (c *xyConnection) writeText(ctx context.Context, payload []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	stop := context.AfterFunc(ctx, func() { _ = c.conn.SetWriteDeadline(time.Now()) })
	defer stop()
	if deadline, ok := ctx.Deadline(); ok {
		_ = c.conn.SetWriteDeadline(deadline)
	}
	err := websocket.Message.Send(c.conn, string(payload))
	_ = c.conn.SetWriteDeadline(time.Time{})
	if err != nil && context.Cause(ctx) != nil {
		return context.Cause(ctx)
	}
	return err
}

var webSocketMIDSequence atomic.Uint64

func webSocketMID() string {
	return strconv.FormatInt(time.Now().UnixMilli(), 10) + strconv.FormatUint(webSocketMIDSequence.Add(1), 10) + " 0"
}
