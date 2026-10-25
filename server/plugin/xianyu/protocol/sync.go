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
	SyncPara   = "/s/para"
	SyncPush   = "/s/sync"
	SyncVulcan = "/s/vulcan"

	maxEnvelopeBytes   = 2 << 20
	maxSyncItems       = 100
	maxEncodedItemSize = 1 << 20
)

var (
	ErrMalformedSync = errors.New("闲鱼同步消息格式错误")
	msgpackHandle    = newMsgpackHandle()
)

// ChatMessage 是协议层输出，不直接映射数据库模型。
type ChatMessage struct {
	LWP              string
	PNMID            string
	SessionID        string
	ContentType      int
	Content          string
	SenderUserName   string
	SenderUserID     string
	SenderAppVersion string
	SenderPlatform   string
	ReminderURL      string
	ProductID        string
	OrderID          string
	MessageTime      int64
}

type syncEnvelope struct {
	LWP  string `json:"lwp"`
	Body struct {
		SyncPushPackage *struct {
			Data []struct {
				Data string `json:"data"`
			} `json:"data"`
		} `json:"syncPushPackage"`
	} `json:"body"`
}

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

func contentCandidates(raw string) [][]byte {
	candidates := [][]byte{[]byte(raw)}
	if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil {
		candidates = append(candidates, decoded)
	} else if decoded, err := base64.RawStdEncoding.DecodeString(raw); err == nil {
		candidates = append(candidates, decoded)
	}
	return candidates
}

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

func isSyncCursor(payload map[string]any) bool {
	_, hasUser := payload["1"].(string)
	_, hasCursor := payload["3"].(string)
	_, hasTime := integer(payload["4"])
	return hasUser && hasCursor && hasTime
}

func stringField(values map[string]any, key string) (string, bool) {
	value, ok := values[key].(string)
	return value, ok
}

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

func senderName(title, content string) string {
	if name, _, ok := strings.Cut(content, " 向你发送了一条新消息"); ok && name != "" {
		return name
	}
	if name, _, ok := strings.Cut(content, " ："); ok && name != "" && len([]rune(name)) < 20 {
		return name
	}
	return title
}

func orderID(content map[string]any) string {
	if target, ok := nestedString(content, "dxCard", "item", "main", "exContent", "button", "targetUrl"); ok {
		if id := queryValue(target, "orderId", "id"); id != "" {
			return id
		}
	}
	target, _ := nestedString(content, "dynamicOperation", "changeContent", "dxCard", "item", "main", "exContent", "button", "targetUrl")
	return queryValue(target, "orderId", "id")
}

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
