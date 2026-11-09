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

// Dial 获取账号消息令牌和 Cookie，建立并注册闲鱼 WebSocket 连接。
func (d *XYDialer) Dial(ctx context.Context, accountID uint) (connection.Connection, error) {
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
	connection := &XYConnection{accountID: accountID, conn: conn, pending: make(map[string]chan WebSocketResponse)}
	registerCtx, cancel := context.WithTimeoutCause(ctx, 10*time.Second, errors.New("闲鱼 WebSocket 注册超时"))
	defer cancel()
	if err := connection.register(registerCtx, token, deviceID); err != nil {
		_ = connection.Close()
		return nil, err
	}
	return connection, nil
}

// register 发送客户端能力并等待闲鱼返回有效注册会话。
func (c *XYConnection) register(ctx context.Context, token, deviceID string) error {
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
			Code    int            `json:"code"`    // Code 是注册响应状态码。
			Headers map[string]any `json:"headers"` // Headers 包含注册会话标识。
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

// Read 持续读取协议帧，路由请求响应、确认推送并返回业务消息。
func (c *XYConnection) Read(ctx context.Context) ([]byte, error) {
	for {
		message, err := c.receive(ctx)
		if err != nil {
			return nil, err
		}
		var envelope struct {
			LWP     string          `json:"lwp"`     // LWP 是协议方法名。
			Code    int             `json:"code"`    // Code 是协议状态码。
			Headers map[string]any  `json:"headers"` // Headers 包含消息关联字段。
			Body    json.RawMessage `json:"body"`    // Body 是协议业务负载。
		}
		if err := json.Unmarshal(message, &envelope); err != nil {
			zap.L().Warn("闲鱼 WebSocket 收到无法解析的协议帧", zap.Uint("accountID", c.accountID), zap.Int("bytes", len(message)), zap.Error(err))
			continue
		}
		var pushBody struct {
			SyncPushPackage json.RawMessage `json:"syncPushPackage"` // SyncPushPackage 是消息同步包。
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
			c.resolve(mid, WebSocketResponse{code: code, raw: message})
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

// hasPending 判断消息 ID 是否对应一个等待中的请求。
func (c *XYConnection) hasPending(mid string) bool {
	c.pendingMu.Lock()
	defer c.pendingMu.Unlock()
	return c.pending[mid] != nil
}

// Send 发送请求并等待成功响应，忽略响应正文。
func (c *XYConnection) Send(ctx context.Context, payload []byte) error {
	_, err := c.Request(ctx, payload)
	return err
}

// Request 按 payload 中的消息 ID 关联并等待对应响应。
func (c *XYConnection) Request(ctx context.Context, payload []byte) ([]byte, error) {
	var envelope struct {
		Headers struct {
			MID string `json:"mid"` // MID 是请求关联标识。
		} `json:"headers"` // Headers 包含协议请求头。
	}
	if err := json.Unmarshal(payload, &envelope); err != nil || envelope.Headers.MID == "" {
		return nil, errors.New("闲鱼发送消息缺少请求标识")
	}
	result := make(chan WebSocketResponse, 1)
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

// resolve 非阻塞地把 response 交付给指定消息 ID 的等待者。
func (c *XYConnection) resolve(mid string, response WebSocketResponse) {
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

// Ping 发送闲鱼协议心跳帧。
func (c *XYConnection) Ping(ctx context.Context) error {
	return c.writeJSON(ctx, map[string]any{"lwp": "/!", "headers": map[string]any{"mid": webSocketMID()}})
}

// Close 幂等关闭底层 WebSocket 连接。
func (c *XYConnection) Close() error {
	var err error
	c.closeOnce.Do(func() { err = c.conn.Close() })
	return err
}

// ack 回传协议要求的关联请求头和成功状态。
func (c *XYConnection) ack(ctx context.Context, headers map[string]any) error {
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

// receive 将上下文取消或截止时间映射为 WebSocket 读截止时间。
func (c *XYConnection) receive(ctx context.Context) ([]byte, error) {
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

// writeJSON 串行发送 JSON 帧并遵守上下文截止时间。
func (c *XYConnection) writeJSON(ctx context.Context, payload any) error {
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

// writeText 串行发送文本帧并遵守上下文截止时间。
func (c *XYConnection) writeText(ctx context.Context, payload []byte) error {
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

// webSocketMIDSequence 避免同毫秒生成重复消息 ID。
var webSocketMIDSequence atomic.Uint64

// webSocketMID 生成闲鱼协议要求的近似单调消息 ID。
func webSocketMID() string {
	return strconv.FormatInt(time.Now().UnixMilli(), 10) + strconv.FormatUint(webSocketMIDSequence.Add(1), 10) + " 0"
}
