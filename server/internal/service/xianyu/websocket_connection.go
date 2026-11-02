package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/gcc798/xianyu.bot/internal/xianyu/connection"
	"github.com/gcc798/xianyu.bot/internal/xianyu/protocol"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

func (d *xyDialer) Dial(ctx context.Context, accountID uint) (connection.Connection, error) {
	var account model.Account
	if err := d.db.WithContext(ctx).Where("id = ?", accountID).First(&account).Error; err != nil {
		return nil, err
	}
	token, deviceID, err := d.token.Get(ctx, account.OwnerID, accountID)
	if err != nil {
		return nil, err
	}
	var credential model.Credential
	if err := d.db.WithContext(ctx).Where("account_id = ?", accountID).First(&credential).Error; err != nil {
		return nil, err
	}
	cookie, err := decryptCredential(d.db, credential.CookieCiphertext, credentialAAD(accountID, "cookie"))
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
			zap.L().Warn("闲鱼 WebSocket 收到无法解析的协议帧", zap.Uint("accountID", c.accountID), zap.Int("bytes", len(message)), zap.Error(err))
			continue
		}
		var pushBody struct {
			SyncPushPackage json.RawMessage `json:"syncPushPackage"`
		}
		_ = json.Unmarshal(envelope.Body, &pushBody)
		if envelope.LWP != "" && envelope.LWP != "/!" {
			zap.L().Info("闲鱼 WebSocket 收到协议帧",
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
