package protocol

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ParseConversationResponse extracts single-chat conversation IDs from an IM RPC response.
func ParseConversationResponse(raw []byte) ([]string, error) {
	var response struct {
		Body struct {
			Reason    string `json:"reason"`
			UserConvs []struct {
				Single struct {
					Conversation struct {
						CID string `json:"cid"`
					} `json:"singleChatConversation"`
				} `json:"singleChatUserConversation"`
			} `json:"userConvs"`
		} `json:"body"`
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

// ParseHistoryResponse converts listUserMessages output into the same events as live pushes.
func ParseHistoryResponse(raw []byte, sessionID string) ([]ChatMessage, error) {
	var response struct {
		Body struct {
			Reason            string            `json:"reason"`
			UserMessageModels []json.RawMessage `json:"userMessageModels"`
		} `json:"body"`
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

func parseHistoryMessage(raw []byte, sessionID string) (ChatMessage, bool) {
	var model struct {
		Message struct {
			MessageID any             `json:"messageId"`
			CreateAt  any             `json:"createAt"`
			Time      any             `json:"time"`
			Extension json.RawMessage `json:"extension"`
			Content   struct {
				Custom struct {
					Data    string `json:"data"`
					Summary string `json:"summary"`
				} `json:"custom"`
			} `json:"content"`
		} `json:"message"`
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

func decodeContentString(raw string) (map[string]any, bool) {
	for _, candidate := range contentCandidates(raw) {
		var content map[string]any
		if json.Unmarshal(candidate, &content) == nil {
			return content, true
		}
	}
	return nil, false
}

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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
