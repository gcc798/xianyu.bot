package protocol

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ParseConversationResponse 从 IM RPC 响应提取去重后的单聊会话标识。
func ParseConversationResponse(raw []byte) ([]string, error) {
	var response struct {
		Body struct {
			Reason    string `json:"reason"` // Reason 是 RPC 失败原因。
			UserConvs []struct {
				Single struct {
					Conversation struct {
						CID string `json:"cid"` // CID 是单聊会话标识。
					} `json:"singleChatConversation"` // Conversation 保存单聊标识。
				} `json:"singleChatUserConversation"` // Single 保存单聊会话数据。
			} `json:"userConvs"` // UserConvs 是用户最近会话列表。
		} `json:"body"` // Body 是 RPC 业务结果。
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, fmt.Errorf("解析闲鱼会话响应: %w", err)
	}
	if response.Body.Reason != "" {
		return nil, errors.New(response.Body.Reason)
	}
	result := make([]string, 0, len(response.Body.UserConvs))
	seen := make(map[string]struct{}, len(response.Body.UserConvs))
	for _, item := range response.Body.UserConvs {
		cid := strings.TrimSpace(item.Single.Conversation.CID)
		if cid == "" {
			continue
		}
		if _, exists := seen[cid]; exists {
			continue
		}
		seen[cid] = struct{}{}
		result = append(result, cid)
	}
	return result, nil
}

// ParseHistoryResponse 将历史消息响应转换为与实时推送一致的聊天事件。
func ParseHistoryResponse(raw []byte, sessionID string) ([]ChatMessage, error) {
	var response struct {
		Body struct {
			Reason            string            `json:"reason"`            // Reason 是 RPC 失败原因。
			UserMessageModels []json.RawMessage `json:"userMessageModels"` // UserMessageModels 是原始历史消息列表。
		} `json:"body"` // Body 是 RPC 业务结果。
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, fmt.Errorf("解析闲鱼消息响应: %w", err)
	}
	if response.Body.Reason != "" {
		return nil, errors.New(response.Body.Reason)
	}
	messages := make([]ChatMessage, 0, len(response.Body.UserMessageModels))
	for _, rawModel := range response.Body.UserMessageModels {
		message, ok := parseHistoryMessage(rawModel, sessionID)
		if ok {
			messages = append(messages, message)
		}
	}
	return messages, nil
}

// parseHistoryMessage 将单条历史消息归一化，缺少会话标识时丢弃。
func parseHistoryMessage(raw []byte, sessionID string) (ChatMessage, bool) {
	var model struct {
		Message struct {
			MessageID any             `json:"messageId"` // MessageID 是可能漂移类型的消息标识。
			CreateAt  any             `json:"createAt"`  // CreateAt 是首选消息时间。
			Time      any             `json:"time"`      // Time 是兼容旧响应的消息时间。
			Extension json.RawMessage `json:"extension"` // Extension 保存发送方和业务关联信息。
			Content   struct {
				Custom struct {
					Data    string `json:"data"`    // Data 是编码后的实际消息内容。
					Summary string `json:"summary"` // Summary 是内容解析失败时的摘要。
				} `json:"custom"` // Custom 保存闲鱼自定义消息体。
			} `json:"content"` // Content 保存协议消息内容。
		} `json:"message"` // Message 是历史消息主体。
	}
	if json.Unmarshal(raw, &model) != nil {
		return ChatMessage{}, false
	}
	extension := rawObject(model.Message.Extension)
	message := ChatMessage{LWP: "/r/MessageManager/listUserMessages", SessionID: sessionID}
	message.PNMID = strings.TrimSpace(fmt.Sprint(model.Message.MessageID))
	if message.PNMID == "" || message.PNMID == "<nil>" {
		message.PNMID = fmt.Sprintf("%x", sha256.Sum256(raw))
	}
	message.MessageTime, _ = integer(model.Message.CreateAt)
	if message.MessageTime == 0 {
		message.MessageTime, _ = integer(model.Message.Time)
	}
	message.SenderUserID, _ = stringField(extension, "senderUserId")
	message.SenderUserName, _ = stringField(extension, "senderNick")
	if message.SenderUserName == "" {
		message.SenderUserName, _ = stringField(extension, "reminderTitle")
	}
	message.ReminderURL, _ = stringField(extension, "reminderUrl")
	message.ProductID = queryValue(message.ReminderURL, "itemId")
	message.OrderID = queryValue(message.ReminderURL, "orderId")
	if content, ok := decodeContentString(model.Message.Content.Custom.Data); ok {
		if value, ok := integer(content["contentType"]); ok {
			message.ContentType = int(value)
		}
		message.Content = contentText(content)
		if images := contentImages(content); message.Content == "" && len(images) > 0 {
			message.Content = strings.Join(images, "\n")
		}
		if message.ContentType == 3 && message.Content == "" {
			message.Content = "[语音消息]"
		}
		message.OrderID = firstNonEmpty(message.OrderID, orderID(content))
	}
	if message.Content == "" {
		message.Content = model.Message.Content.Custom.Summary
	}
	if message.ProductID == "" {
		message.ProductID = embeddedID(extension, "itemId")
	}
	if message.OrderID == "" {
		message.OrderID = embeddedID(extension, "orderId")
	}
	return message, message.SessionID != ""
}

// rawObject 兼容直接对象和 JSON 字符串两种扩展字段。
func rawObject(raw json.RawMessage) map[string]any {
	var result map[string]any
	if json.Unmarshal(raw, &result) == nil {
		return result
	}
	var encoded string
	if json.Unmarshal(raw, &encoded) == nil {
		_ = json.Unmarshal([]byte(encoded), &result)
	}
	return result
}

// decodeContentString 从原文或 Base64 候选中解码消息对象。
func decodeContentString(raw string) (map[string]any, bool) {
	for _, candidate := range contentCandidates(raw) {
		var content map[string]any
		if json.Unmarshal(candidate, &content) == nil {
			return content, true
		}
	}
	return nil, false
}

// embeddedID 从扩展 JSON 的已知字段中读取业务标识。
func embeddedID(extension map[string]any, key string) string {
	for _, field := range []string{"bizTag", "extJson"} {
		raw, _ := stringField(extension, field)
		var values map[string]any
		if json.Unmarshal([]byte(raw), &values) == nil {
			if value := strings.TrimSpace(fmt.Sprint(values[key])); value != "" && value != "<nil>" {
				return value
			}
		}
	}
	return ""
}

// firstNonEmpty 返回第一个非空字符串。
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
