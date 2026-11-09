package protocol

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/ugorji/go/codec"
)

const (
	// SyncPara 是参数同步协议方法。
	SyncPara = "/s/para"
	// SyncPush 是消息同步推送协议方法。
	SyncPush = "/s/sync"
	// SyncVulcan 是 Vulcan 同步协议方法。
	SyncVulcan = "/s/vulcan"

	// maxEnvelopeBytes 是单个同步外层帧大小上限。
	maxEnvelopeBytes = 2 << 20
	// maxSyncItems 是单帧允许的同步条目上限。
	maxSyncItems = 100
	// maxEncodedItemSize 是单条 Base64 数据大小上限。
	maxEncodedItemSize = 1 << 20
)

var (
	// ErrMalformedSync 表示同步帧不满足协议或安全边界。
	ErrMalformedSync = errors.New("闲鱼同步消息格式错误")
	// msgpackHandle 是协议 MessagePack 解码配置模板。
	msgpackHandle = newMsgpackHandle()
)

// ChatMessage 是实时与历史协议统一输出的聊天事件。
type ChatMessage struct {
	LWP              string // LWP 是来源协议方法。
	PNMID            string // PNMID 是消息去重标识。
	SessionID        string // SessionID 是闲鱼会话标识。
	ContentType      int    // ContentType 是闲鱼内容类型编码。
	Content          string // Content 是归一化文本或图片地址。
	SenderUserName   string // SenderUserName 是发送方显示名称。
	SenderUserID     string // SenderUserID 是闲鱼发送方标识。
	SenderAppVersion string // SenderAppVersion 是发送方客户端版本。
	SenderPlatform   string // SenderPlatform 是发送方平台。
	ReminderURL      string // ReminderURL 是消息携带的提醒链接。
	ProductID        string // ProductID 是关联的闲鱼商品标识。
	OrderID          string // OrderID 是关联的闲鱼订单标识。
	MessageTime      int64  // MessageTime 是闲鱼侧毫秒时间戳。
}

// syncEnvelope 是闲鱼同步推送的最小 JSON 外层结构。
type syncEnvelope struct {
	LWP  string `json:"lwp"` // LWP 是同步协议方法。
	Body struct {
		SyncPushPackage *struct {
			Data []struct {
				Data string `json:"data"` // Data 是 Base64 编码的同步条目。
			} `json:"data"` // Data 是当前推送包含的条目列表。
		} `json:"syncPushPackage"` // SyncPushPackage 是消息同步包。
	} `json:"body"` // Body 是协议业务负载。
}

// newMsgpackHandle 创建带深度和初始容量限制的安全解码配置。
func newMsgpackHandle() *codec.MsgpackHandle {
	handle := new(codec.MsgpackHandle)
	handle.MapType = reflect.TypeFor[map[string]any]()
	handle.SignedInteger = true
	handle.RawToString = true
	handle.MaxDepth = 32
	handle.MaxInitLen = 1024
	return handle
}

// ParseSyncEnvelope 将同步推送解析为内部聊天事件；已读回执会被忽略。
func ParseSyncEnvelope(raw []byte) ([]ChatMessage, error) {
	if len(raw) == 0 || len(raw) > maxEnvelopeBytes {
		return nil, fmt.Errorf("%w: 消息体为空或过大", ErrMalformedSync)
	}
	var envelope syncEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("%w: 外层 JSON: %v", ErrMalformedSync, err)
	}
	if envelope.Body.SyncPushPackage == nil || envelope.Body.SyncPushPackage.Data == nil {
		return nil, fmt.Errorf("%w: 缺少同步数据", ErrMalformedSync)
	}
	items := envelope.Body.SyncPushPackage.Data
	if len(items) > maxSyncItems {
		return nil, fmt.Errorf("%w: 同步数据条数无效", ErrMalformedSync)
	}

	messages := make([]ChatMessage, 0, len(items))
	var firstErr error
	for index, item := range items {
		message, keep, err := parseSyncItem(envelope.LWP, item.Data)
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("%w: 第 %d 条: %v", ErrMalformedSync, index+1, err)
			}
			continue
		}
		if keep {
			messages = append(messages, message)
		}
	}
	if len(messages) == 0 && firstErr != nil {
		return nil, firstErr
	}
	return messages, nil
}

// parseSyncItem 解码单条 JSON 或 MessagePack 数据，并过滤回执与游标。
func parseSyncItem(lwp, encoded string) (ChatMessage, bool, error) {
	if encoded == "" || len(encoded) > maxEncodedItemSize {
		return ChatMessage{}, false, errors.New("MessagePack 数据为空或过大")
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		raw, err = base64.RawStdEncoding.DecodeString(encoded)
	}
	if err != nil {
		return ChatMessage{}, false, fmt.Errorf("Base64 解码失败: %w", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		decoded, bytesRead, err := decodeMsgpackMap(raw)
		if err != nil {
			return ChatMessage{}, false, fmt.Errorf("JSON/MessagePack 解码失败: %w", err)
		}
		if bytesRead != len(raw) {
			return ChatMessage{}, false, errors.New("MessagePack 包含多余数据")
		}
		payload = decoded
	}
	messageType, hasMessageType := integer(payload["2"])
	if hasMessageType && messageType == 2 {
		return ChatMessage{}, false, nil
	}
	if hasMessageType && messageType == 1 && isSyncCursor(payload) {
		return ChatMessage{}, false, nil
	}

	messageInfo, ok := payload["1"].(map[string]any)
	if !ok {
		return parseCardUpdate(lwp, payload, raw)
	}
	message := ChatMessage{LWP: lwp}
	message.SessionID, _ = stringField(messageInfo, "2")
	message.MessageTime, _ = integer(messageInfo["5"])

	if metadata, ok := messageInfo["10"].(map[string]any); ok {
		message.PNMID = messageID(messageInfo, metadata, raw)
		message.Content, _ = stringField(metadata, "reminderContent")
		title, _ := stringField(metadata, "reminderTitle")
		message.SenderUserName = senderName(title, message.Content)
		message.SenderUserID, _ = stringField(metadata, "senderUserId")
		message.SenderAppVersion, _ = stringField(metadata, "_appVersion")
		message.SenderPlatform, _ = stringField(metadata, "_platform")
		message.ReminderURL, _ = stringField(metadata, "reminderUrl")
		message.ProductID = queryValue(message.ReminderURL, "itemId")
	}

	content, found := decodeMessageContent(messageInfo)
	if found {
		if value, ok := integer(content["contentType"]); ok {
			message.ContentType = int(value)
		}
		if exact := contentText(content); exact != "" {
			message.Content = exact
		} else if images := contentImages(content); len(images) > 0 {
			message.Content = strings.Join(images, "\n")
			if message.ContentType == 0 {
				message.ContentType = 2
			}
		} else if message.ContentType == 3 {
			message.Content = "[语音消息]"
		}
		message.OrderID = orderID(content)
	}
	if message.PNMID == "" {
		message.PNMID = fmt.Sprintf("%x", sha256.Sum256(raw))
	}

	return message, true, nil
}

// decodeMessageContent 从协议候选路径中解析实际消息内容。
func decodeMessageContent(messageInfo map[string]any) (map[string]any, bool) {
	for _, key := range []string{"5", "1"} {
		raw, ok := nestedString(messageInfo, "6", "3", key)
		if !ok || strings.TrimSpace(raw) == "" {
			continue
		}
		for _, candidate := range contentCandidates(raw) {
			var content map[string]any
			if json.Unmarshal(candidate, &content) == nil && looksLikeContent(content) {
				return content, true
			}
		}
	}
	return nil, false
}

// contentCandidates 返回原文及可成功 Base64 解码的内容候选。
func contentCandidates(raw string) [][]byte {
	candidates := [][]byte{[]byte(raw)}
	if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil {
		candidates = append(candidates, decoded)
	} else if decoded, err := base64.RawStdEncoding.DecodeString(raw); err == nil {
		candidates = append(candidates, decoded)
	}
	return candidates
}

// looksLikeContent 判断对象是否包含已知消息内容字段。
func looksLikeContent(content map[string]any) bool {
	if _, ok := content["contentType"]; ok {
		return true
	}
	for _, key := range []string{"text", "image", "picUrl", "audio", "dxCard", "dynamicOperation"} {
		if _, ok := content[key]; ok {
			return true
		}
	}
	return false
}

// contentText 兼容字符串和对象两种文本结构。
func contentText(content map[string]any) string {
	switch text := content["text"].(type) {
	case string:
		return text
	case map[string]any:
		value, _ := stringField(text, "text")
		return value
	default:
		return ""
	}
}

// contentImages 提取单图或多图消息中的图片地址。
func contentImages(content map[string]any) []string {
	if raw, ok := content["picUrl"].(string); ok && raw != "" {
		return []string{raw}
	}
	image, ok := content["image"].(map[string]any)
	if !ok {
		return nil
	}
	pics, ok := image["pics"].([]any)
	if !ok {
		return nil
	}
	urls := make([]string, 0, len(pics))
	for _, raw := range pics {
		pic, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if value, ok := stringField(pic, "url"); ok && value != "" {
			urls = append(urls, value)
		}
	}
	return urls
}

// decodeMsgpackMap 解码顶层对象并返回实际消费字节数。
func decodeMsgpackMap(raw []byte) (map[string]any, int, error) {
	handle := newMsgpackHandle()
	handle.MapType = reflect.TypeFor[map[any]any]()
	var decoded any
	decoder := codec.NewDecoderBytes(raw, handle)
	if err := decoder.Decode(&decoded); err != nil {
		return nil, decoder.NumBytesRead(), err
	}
	payload, ok := normalizeMsgpack(decoded).(map[string]any)
	if !ok {
		return nil, decoder.NumBytesRead(), errors.New("MessagePack 顶层不是对象")
	}
	return payload, decoder.NumBytesRead(), nil
}

// normalizeMsgpack 递归把 MessagePack 任意键对象转换为字符串键对象。
func normalizeMsgpack(value any) any {
	switch current := value.(type) {
	case map[any]any:
		result := make(map[string]any, len(current))
		for key, child := range current {
			result[fmt.Sprint(key)] = normalizeMsgpack(child)
		}
		return result
	case []any:
		for index, child := range current {
			current[index] = normalizeMsgpack(child)
		}
	}
	return value
}

// parseCardUpdate 将无标准消息体的订单卡片更新归一为聊天事件。
func parseCardUpdate(lwp string, payload map[string]any, raw []byte) (ChatMessage, bool, error) {
	metadata, ok := payload["4"].(map[string]any)
	if !ok {
		return ChatMessage{}, false, nil
	}
	message := ChatMessage{LWP: lwp, PNMID: messageID(nil, metadata, raw)}
	message.SessionID, _ = payload["2"].(string)
	message.MessageTime, _ = integer(payload["5"])
	message.Content, _ = stringField(metadata, "reminderContent")
	message.SenderUserID, _ = stringField(metadata, "senderUserId")
	message.SenderUserName, _ = stringField(metadata, "senderNick")
	if message.SenderUserName == "" {
		message.SenderUserName, _ = stringField(metadata, "reminderTitle")
	}
	message.ReminderURL, _ = stringField(metadata, "reminderUrl")
	message.ProductID = queryValue(message.ReminderURL, "itemId")
	message.OrderID = queryValue(message.ReminderURL, "orderId")
	if strings.Contains(message.Content, "已付款") || strings.Contains(message.Content, "等待你发货") {
		message.ContentType = 26
	}
	return message, true, nil
}

// messageID 按显式字段、内嵌字段、内容摘要的顺序生成去重标识。
func messageID(messageInfo, metadata map[string]any, raw []byte) string {
	if messageInfo != nil {
		if id, ok := stringField(messageInfo, "3"); ok && id != "" {
			return id
		}
	}
	for _, field := range []string{"bizTag", "extJson"} {
		embedded, _ := stringField(metadata, field)
		var value map[string]any
		if json.Unmarshal([]byte(embedded), &value) == nil {
			if id, ok := value["messageId"].(string); ok && id != "" {
				return id
			}
		}
	}
	return fmt.Sprintf("%x", sha256.Sum256(raw))
}

// isSyncCursor 判断同步条目是否仅用于推进游标。
func isSyncCursor(payload map[string]any) bool {
	_, hasUser := payload["1"].(string)
	_, hasCursor := payload["3"].(string)
	_, hasTime := integer(payload["4"])
	return hasUser && hasCursor && hasTime
}

// stringField 从对象读取严格字符串字段。
func stringField(values map[string]any, key string) (string, bool) {
	value, ok := values[key].(string)
	return value, ok
}

// integer 将协议中的常见整数表示安全转换为 int64。
func integer(value any) (int64, bool) {
	switch number := value.(type) {
	case int64:
		return number, true
	case uint64:
		if number <= uint64(^uint64(0)>>1) {
			return int64(number), true
		}
	case int:
		return int64(number), true
	case float64:
		return int64(number), number == float64(int64(number))
	}
	return 0, false
}

// nestedString 沿 path 读取嵌套对象中的字符串。
func nestedString(root map[string]any, path ...string) (string, bool) {
	var current any = root
	for _, key := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return "", false
		}
		current, ok = object[key]
		if !ok {
			return "", false
		}
	}
	value, ok := current.(string)
	return value, ok
}

// senderName 优先从提醒正文提取发送方名称，失败时使用标题。
func senderName(title, content string) string {
	if name, _, ok := strings.Cut(content, " 向你发送了一条新消息"); ok && name != "" {
		return name
	}
	if name, _, ok := strings.Cut(content, " ："); ok && name != "" && len([]rune(name)) < 20 {
		return name
	}
	return title
}

// orderID 从已知卡片跳转路径中提取订单标识。
func orderID(content map[string]any) string {
	if target, ok := nestedString(content, "dxCard", "item", "main", "exContent", "button", "targetUrl"); ok {
		if id := queryValue(target, "orderId", "id"); id != "" {
			return id
		}
	}
	target, _ := nestedString(content, "dynamicOperation", "changeContent", "dxCard", "item", "main", "exContent", "button", "targetUrl")
	return queryValue(target, "orderId", "id")
}

// queryValue 按 keys 顺序返回 URL 中第一个非空查询值。
func queryValue(rawURL string, keys ...string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	query := parsed.Query()
	for _, key := range keys {
		if value := query.Get(key); value != "" {
			return value
		}
	}
	return ""
}
