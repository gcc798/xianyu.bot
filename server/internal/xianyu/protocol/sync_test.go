package protocol

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ugorji/go/codec"
)

// TestParseSyncEnvelopeFixture 验证对应测试场景。
func TestParseSyncEnvelopeFixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/sync_message.json")
	require.NoError(t, err)

	messages, err := ParseSyncEnvelope(raw)
	require.NoError(t, err)
	require.Len(t, messages, 1)
	require.Equal(t, ChatMessage{
		LWP:              SyncPush,
		PNMID:            "message-demo.PNM",
		SessionID:        "session-demo@goofish",
		ContentType:      26,
		Content:          "测试买家 ：你好",
		SenderUserName:   "测试买家",
		SenderUserID:     "buyer-demo",
		SenderAppVersion: "1.0.0-test",
		SenderPlatform:   "test",
		ReminderURL:      "fleamarket://item/detail?itemId=item-demo",
		ProductID:        "item-demo",
		OrderID:          "order-demo",
		MessageTime:      1761091200000,
	}, messages[0])
}

// TestParseSyncEnvelopeSupportsCurrentLWPsAndIgnoresReceipt 验证对应测试场景。
func TestParseSyncEnvelopeSupportsCurrentLWPsAndIgnoresReceipt(t *testing.T) {
	encoded := encodePayload(t, map[string]any{"1": "session-demo@goofish", "2": int64(2)})
	for _, lwp := range []string{SyncPara, SyncVulcan} {
		t.Run(lwp, func(t *testing.T) {
			messages, err := ParseSyncEnvelope(envelopeJSON(t, lwp, encoded))
			require.NoError(t, err)
			require.Empty(t, messages)
		})
	}
}

// TestParseSyncEnvelopeIgnoresSyncCursor 验证对应测试场景。
func TestParseSyncEnvelopeIgnoresSyncCursor(t *testing.T) {
	encoded := encodePayload(t, map[string]any{
		"1": "user-demo@goofish",
		"2": int64(1),
		"3": "cursor-demo",
		"4": int64(1761091200000),
	})

	messages, err := ParseSyncEnvelope(envelopeJSON(t, SyncPush, encoded))
	require.NoError(t, err)
	require.Empty(t, messages)
}

// TestParseSyncEnvelopeAllowsEmptyBatch 验证对应测试场景。
func TestParseSyncEnvelopeAllowsEmptyBatch(t *testing.T) {
	raw := []byte(`{"lwp":"/s/sync","body":{"syncPushPackage":{"data":[]}}}`)
	messages, err := ParseSyncEnvelope(raw)
	require.NoError(t, err)
	require.Empty(t, messages)
}

// TestParseSyncEnvelopeUsesMetadataMessageIDAndExactText 验证对应测试场景。
func TestParseSyncEnvelopeUsesMetadataMessageIDAndExactText(t *testing.T) {
	payload := map[string]any{"1": map[string]any{
		"2": "session@goofish", "5": int64(123),
		"10": map[string]any{"reminderContent": "占位文本", "senderUserId": "buyer", "extJson": `{"messageId":"message-from-meta"}`},
		"6":  map[string]any{"3": map[string]any{"5": `{"contentType":1,"text":{"text":"价格是什么"}}`}},
	}}
	messages, err := ParseSyncEnvelope(envelopeJSON(t, "/s/new", encodePayload(t, payload)))
	require.NoError(t, err)
	require.Len(t, messages, 1)
	require.Equal(t, "message-from-meta", messages[0].PNMID)
	require.Equal(t, "价格是什么", messages[0].Content)
}

// TestParseSyncEnvelopeSupportsBase64ContentFallbackAndImages 验证对应测试场景。
func TestParseSyncEnvelopeSupportsBase64ContentFallbackAndImages(t *testing.T) {
	content := base64.StdEncoding.EncodeToString([]byte(`{"contentType":2,"image":{"pics":[{"url":"https://img/1.jpg"},{"url":"https://img/2.jpg"}]}}`))
	payload := map[string]any{"1": map[string]any{
		"2": "session@goofish", "3": "message-image", "5": int64(123),
		"10": map[string]any{"reminderContent": "[图片]", "senderUserId": "buyer"},
		"6":  map[string]any{"3": map[string]any{"1": content}},
	}}
	messages, err := ParseSyncEnvelope(envelopeJSON(t, SyncPush, encodePayload(t, payload)))
	require.NoError(t, err)
	require.Len(t, messages, 1)
	require.Equal(t, 2, messages[0].ContentType)
	require.Equal(t, "https://img/1.jpg\nhttps://img/2.jpg", messages[0].Content)
}

// TestParseSyncEnvelopeSupportsBase64JSONPayload 验证对应测试场景。
func TestParseSyncEnvelopeSupportsBase64JSONPayload(t *testing.T) {
	payload := map[string]any{"1": map[string]any{
		"2": "session@goofish", "3": "message-json", "5": 123,
		"10": map[string]any{"reminderContent": "联调2254：价格是什么", "senderUserId": "buyer"},
	}}
	raw, err := json.Marshal(payload)
	require.NoError(t, err)

	messages, err := ParseSyncEnvelope(envelopeJSON(t, SyncPara, base64.StdEncoding.EncodeToString(raw)))
	require.NoError(t, err)
	require.Len(t, messages, 1)
	require.Equal(t, "message-json", messages[0].PNMID)
	require.Equal(t, "联调2254：价格是什么", messages[0].Content)
}

// TestParseSyncEnvelopeSupportsIntegerMessagePackKeys 验证对应测试场景。
func TestParseSyncEnvelopeSupportsIntegerMessagePackKeys(t *testing.T) {
	payload := map[any]any{1: map[any]any{
		2: "session@goofish", 3: "message-integer-keys", 5: int64(123),
		10: map[any]any{"reminderContent": "价格是什么", "senderUserId": "buyer"},
	}}
	var raw []byte
	require.NoError(t, codec.NewEncoderBytes(&raw, msgpackHandle).Encode(payload))

	messages, err := ParseSyncEnvelope(envelopeJSON(t, SyncPara, base64.RawStdEncoding.EncodeToString(raw)))
	require.NoError(t, err)
	require.Len(t, messages, 1)
	require.Equal(t, "message-integer-keys", messages[0].PNMID)
	require.Equal(t, "价格是什么", messages[0].Content)
}

// TestParseSyncEnvelopeSupportsPaidCardUpdate 验证对应测试场景。
func TestParseSyncEnvelopeSupportsPaidCardUpdate(t *testing.T) {
	payload := map[string]any{
		"1": "item", "2": "session@goofish", "5": int64(123),
		"4": map[string]any{"reminderContent": "我已付款，等待你发货", "senderUserId": "buyer", "senderNick": "买家", "reminderUrl": "https://www.goofish.com/?itemId=item-1&orderId=order-1"},
	}
	messages, err := ParseSyncEnvelope(envelopeJSON(t, "/s/new", encodePayload(t, payload)))
	require.NoError(t, err)
	require.Len(t, messages, 1)
	require.Equal(t, 26, messages[0].ContentType)
	require.Equal(t, "item-1", messages[0].ProductID)
	require.Equal(t, "order-1", messages[0].OrderID)
}

// TestParseSyncEnvelopeRejectsMalformedProtocol 验证对应测试场景。
func TestParseSyncEnvelopeRejectsMalformedProtocol(t *testing.T) {
	tests := []struct {
		name   string
		raw    []byte
		target error
	}{
		{name: "invalid base64", raw: envelopeJSON(t, SyncPush, "%%%"), target: ErrMalformedSync},
		{name: "missing payload", raw: []byte(`{"lwp":"/s/sync","body":{}}`), target: ErrMalformedSync},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseSyncEnvelope(test.raw)
			require.ErrorIs(t, err, test.target)
		})
	}
}

// encodePayload 提供测试辅助能力。
func encodePayload(t *testing.T, payload map[string]any) string {
	t.Helper()
	var raw []byte
	require.NoError(t, codec.NewEncoderBytes(&raw, msgpackHandle).Encode(payload))
	return base64.RawStdEncoding.EncodeToString(raw)
}

// envelopeJSON 提供测试辅助能力。
func envelopeJSON(t *testing.T, lwp, data string) []byte {
	t.Helper()
	value := map[string]any{
		"lwp":  lwp,
		"body": map[string]any{"syncPushPackage": map[string]any{"data": []any{map[string]any{"data": data}}}},
	}
	raw, err := json.Marshal(value)
	require.NoError(t, err)
	return raw
}

// TestParseSyncEnvelopeSkipsMalformedItemWhenBatchHasValidMessage 验证对应测试场景。
func TestParseSyncEnvelopeSkipsMalformedItemWhenBatchHasValidMessage(t *testing.T) {
	raw, err := os.ReadFile("testdata/sync_message.json")
	require.NoError(t, err)
	var envelope map[string]any
	require.NoError(t, json.Unmarshal(raw, &envelope))
	data := envelope["body"].(map[string]any)["syncPushPackage"].(map[string]any)["data"].([]any)
	envelope["body"].(map[string]any)["syncPushPackage"].(map[string]any)["data"] = append(data, map[string]any{"data": "%%%"})
	raw, err = json.Marshal(envelope)
	require.NoError(t, err)

	messages, err := ParseSyncEnvelope(raw)
	require.NoError(t, err)
	require.Len(t, messages, 1)
	require.Equal(t, "message-demo.PNM", messages[0].PNMID)
}
