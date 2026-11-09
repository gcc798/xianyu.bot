package protocol

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestParseConversationResponse 验证对应测试场景。
func TestParseConversationResponse(t *testing.T) {
	raw := []byte(`{"body":{"userConvs":[{"singleChatUserConversation":{"singleChatConversation":{"cid":"session-1@goofish"}}},{"singleChatUserConversation":{"singleChatConversation":{"cid":"session-1@goofish"}}}]}}`)
	conversations, err := ParseConversationResponse(raw)
	require.NoError(t, err)
	require.Equal(t, []string{"session-1@goofish"}, conversations)
}

// TestParseHistoryResponse 验证对应测试场景。
func TestParseHistoryResponse(t *testing.T) {
	content := base64.StdEncoding.EncodeToString([]byte(`{"contentType":1,"text":{"text":"价格是什么"}}`))
	raw, err := json.Marshal(map[string]any{"body": map[string]any{"userMessageModels": []any{map[string]any{"message": map[string]any{
		"messageId": "history-1", "createAt": int64(123),
		"extension": `{"senderUserId":"buyer@goofish","reminderTitle":"买家","reminderUrl":"https://www.goofish.com/?itemId=item-1"}`,
		"content":   map[string]any{"custom": map[string]any{"data": content}},
	}}}}})
	require.NoError(t, err)
	messages, err := ParseHistoryResponse(raw, "session-1@goofish")
	require.NoError(t, err)
	require.Len(t, messages, 1)
	require.Equal(t, "history-1", messages[0].PNMID)
	require.Equal(t, "价格是什么", messages[0].Content)
	require.Equal(t, "item-1", messages[0].ProductID)
}
