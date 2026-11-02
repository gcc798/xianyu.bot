package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model/xianyu"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/xianyu/request"
	"github.com/gcc798/xianyu.bot/internal/xianyu/connection"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/websocket"
	"gorm.io/gorm"
)

type fixedWebSocketToken struct{}

func TestWebSocketMIDIsUnique(t *testing.T) {
	seen := make(map[string]struct{}, 10000)
	for range 10000 {
		mid := webSocketMID()
		if _, exists := seen[mid]; exists {
			t.Fatalf("duplicate WebSocket request ID: %s", mid)
		}
		seen[mid] = struct{}{}
	}
}

func (fixedWebSocketToken) Get(context.Context, uint, uint) (string, string, error) {
	return "websocket-token", "DEVICE-user-demo", nil
}

func TestXyDialerRegistersAndRoutesSyncMessages(t *testing.T) {
	fixture, err := os.ReadFile("../../xianyu/protocol/testdata/sync_message.json")
	require.NoError(t, err)
	var fixtureEnvelope map[string]any
	require.NoError(t, json.Unmarshal(fixture, &fixtureEnvelope))
	fixtureEnvelope["lwp"] = "/s/current-protocol"
	fixtureEnvelope["headers"] = map[string]any{"mid": "message-mid", "sid": "session-id"}
	fixture = mustJSON(t, fixtureEnvelope)
	serverResult := make(chan error, 1)
	server := httptest.NewServer(websocket.Server{
		Handshake: func(_ *websocket.Config, request *http.Request) error {
			require.Equal(t, webBrowserUA, request.UserAgent())
			return nil
		},
		Handler: func(conn *websocket.Conn) {
			var registration map[string]any
			if err := websocket.JSON.Receive(conn, &registration); err != nil {
				serverResult <- err
				return
			}
			registrationJSON, _ := json.Marshal(registration)
			headers := registration["headers"].(map[string]any)
			if registration["lwp"] != "/reg" || headers["ua"] != webSocketRegistrationUA || strings.Contains(string(registrationJSON), "cookie2") {
				serverResult <- errors.New("注册消息格式错误或泄露 Cookie")
				return
			}
			if err := websocket.JSON.Send(conn, map[string]any{
				"code": 200, "headers": map[string]any{"reg-sid": "registered"},
			}); err != nil {
				serverResult <- err
				return
			}
			var syncStatus map[string]any
			if err := websocket.JSON.Receive(conn, &syncStatus); err != nil || syncStatus["lwp"] != "/r/SyncStatus/ackDiff" {
				serverResult <- errors.New("同步状态消息格式错误")
				return
			}
			if err := websocket.JSON.Send(conn, map[string]any{
				"lwp": "/s/para", "headers": map[string]any{"mid": "control-mid"}, "body": map[string]any{"params": []any{}},
			}); err != nil {
				serverResult <- err
				return
			}
			var controlAck map[string]any
			if err := websocket.JSON.Receive(conn, &controlAck); err != nil || controlAck["code"] != float64(200) {
				serverResult <- errors.New("控制消息确认格式错误")
				return
			}
			if err := websocket.Message.Send(conn, fixture); err != nil {
				serverResult <- err
				return
			}
			var ack map[string]any
			if err := websocket.JSON.Receive(conn, &ack); err != nil || ack["code"] != float64(200) {
				serverResult <- errors.New("消息确认格式错误")
				return
			}
			var sent map[string]any
			if err := websocket.JSON.Receive(conn, &sent); err != nil || sent["lwp"] != "/r/MessageSend/sendByReceiverScope" {
				serverResult <- errors.New("发送消息格式错误")
				return
			}
			sentHeaders := sent["headers"].(map[string]any)
			if err := websocket.JSON.Send(conn, map[string]any{"headers": map[string]any{"mid": sentHeaders["mid"]}, "body": map[string]any{"success": true}}); err != nil {
				serverResult <- err
				return
			}
			var responseAck map[string]any
			if err := websocket.JSON.Receive(conn, &responseAck); err != nil || responseAck["code"] != float64(200) {
				serverResult <- errors.New("请求响应确认格式错误")
				return
			}
			serverResult <- nil
		},
	})
	t.Cleanup(server.Close)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Credential{}))
	t.Setenv(credentialKeyEnvironment, base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))
	accountService := &account{db: db}
	created, err := accountService.Create(t.Context(), 10, xyRequest.CreateAccount{Label: "连接账号"})
	require.NoError(t, err)
	require.NoError(t, accountService.SaveCredential(t.Context(), 10, xyRequest.SaveCredential{
		ID: created.ID, Cookie: "unb=user-demo; cookie2=do-not-send-in-register",
	}))

	dialer := &xyDialer{db: db, token: fixedWebSocketToken{}, endpoint: "ws" + strings.TrimPrefix(server.URL, "http")}
	connection, err := dialer.Dial(t.Context(), created.ID)
	require.NoError(t, err)
	t.Cleanup(func() { _ = connection.Close() })
	control, err := connection.Read(t.Context())
	require.NoError(t, err)
	require.Empty(t, control)
	message, err := connection.Read(t.Context())
	require.NoError(t, err)
	require.JSONEq(t, string(fixture), string(message))
	payload, err := buildTextMessage("session-demo@goofish", "buyer-demo", "user-demo", "你好")
	require.NoError(t, err)
	readCtx, cancelRead := context.WithCancel(t.Context())
	defer cancelRead()
	go func() { _, _ = connection.Read(readCtx) }()
	require.NoError(t, connection.Send(t.Context(), payload))
	require.NoError(t, <-serverResult)
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	return data
}

func TestIsSyncChangeSignal(t *testing.T) {
	require.True(t, isSyncChangeSignal([]byte(`{"lwp":"/s/sync","body":{"syncExtraType":{"type":2}}}`)))
	require.False(t, isSyncChangeSignal([]byte(`{"lwp":"/s/vulcan","body":{"syncExtraType":{"type":3}}}`)))
	require.False(t, isSyncChangeSignal([]byte(`not-json`)))
}

func TestWebSocketStatusIsOwnerIsolated(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Credential{}))
	created, err := (&account{db: db}).Create(t.Context(), 10, xyRequest.CreateAccount{Label: "隔离账号"})
	require.NoError(t, err)

	service := newWebSocketService(t.Context(), db, &businessEvent{db: db}, &message{db: db, events: &businessEvent{db: db}}, nil)
	status, err := service.Status(t.Context(), 10, created.ID)
	require.NoError(t, err)
	require.False(t, status.Running)
	_, err = service.Status(t.Context(), 11, created.ID)
	require.Error(t, err)
}

func TestConnectionStateChangedAlertsOnlyWhenUnavailableAtWindowEnd(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:websocket-alert?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.BusinessEvent{}))
	account := model.Account{OwnerID: 10, Label: "告警账号", Status: "ready", Enabled: true}
	require.NoError(t, db.Create(&account).Error)

	subjects := make(chan string, 3)
	service := &webSocketService{
		db:                  db,
		events:              &businessEvent{db: db},
		alertWindows:        make(map[uint]*connectionAlertWindow),
		alertWindowDuration: 100 * time.Millisecond,
		sendNotificationEmail: func(_ string, subject string, _ string) error {
			subjects <- subject
			return nil
		},
	}
	failure := connection.Snapshot{AccountID: account.ID, State: connection.StateReconnecting, LastError: "连接中断"}
	service.connectionStateChanged(failure)
	service.connectionStateChanged(connection.Snapshot{AccountID: account.ID, State: connection.StateConnected})
	service.connectionStateChanged(failure)
	select {
	case subject := <-subjects:
		t.Fatalf("一小时窗口结束前不应发送邮件: %s", subject)
	case <-time.After(30 * time.Millisecond):
	}
	require.Contains(t, receiveSubject(t, subjects), "连接不可用")

	service.connectionStateChanged(connection.Snapshot{AccountID: account.ID, State: connection.StateConnected})
	require.Contains(t, receiveSubject(t, subjects), "连接已恢复")
	require.Eventually(t, func() bool {
		var count int64
		return db.Model(&model.BusinessEvent{}).Where("account_id = ? AND type = ?", account.ID, "websocket_alert").Count(&count).Error == nil && count == 2
	}, time.Second, 10*time.Millisecond)
	var sentEvents []model.BusinessEvent
	require.NoError(t, db.Where("account_id = ? AND type = ?", account.ID, "websocket_alert").Find(&sentEvents).Error)
	require.Contains(t, sentEvents[0].Summary, "不可用 2 次")

	failedService := &webSocketService{db: db, events: &businessEvent{db: db}, sendNotificationEmail: func(string, string, string) error {
		return errors.New("SMTP 不可用")
	}}
	failedService.sendConnectionNotification(failure, false, time.Now().Add(-time.Hour), 3)
	var failedEvent model.BusinessEvent
	require.NoError(t, db.Where("account_id = ? AND type = ? AND status = ?", account.ID, "websocket_alert", "failed").First(&failedEvent).Error)
	require.Contains(t, failedEvent.Summary, "发送闲鱼消息连接故障邮件失败")
}

func TestConnectionStateChangedSuppressesAlertWhenConnectedAtWindowEnd(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:websocket-alert-suppressed?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.BusinessEvent{}))
	account := model.Account{OwnerID: 10, Label: "恢复账号", Status: "ready", Enabled: true}
	require.NoError(t, db.Create(&account).Error)

	subjects := make(chan string, 1)
	service := &webSocketService{
		db:                  db,
		events:              &businessEvent{db: db},
		alertWindows:        make(map[uint]*connectionAlertWindow),
		alertWindowDuration: 30 * time.Millisecond,
		sendNotificationEmail: func(_ string, subject string, _ string) error {
			subjects <- subject
			return nil
		},
	}
	service.connectionStateChanged(connection.Snapshot{AccountID: account.ID, State: connection.StateReconnecting, LastError: "短暂中断"})
	service.connectionStateChanged(connection.Snapshot{AccountID: account.ID, State: connection.StateConnected})
	var event model.BusinessEvent
	require.Eventually(t, func() bool {
		return db.Where("account_id = ? AND type = ?", account.ID, "websocket_alert").First(&event).Error == nil
	}, time.Second, 10*time.Millisecond)
	select {
	case subject := <-subjects:
		t.Fatalf("窗口结束时连接正常不应发送邮件: %s", subject)
	default:
	}
	require.Contains(t, event.Summary, "未发送告警")
}

func receiveSubject(t *testing.T, subjects <-chan string) string {
	t.Helper()
	select {
	case subject := <-subjects:
		return subject
	case <-time.After(time.Second):
		t.Fatal("等待连接通知邮件超时")
		return ""
	}
}
