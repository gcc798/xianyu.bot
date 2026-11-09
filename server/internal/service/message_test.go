package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gcc798/xianyu.bot/internal/xianyu/protocol"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"github.com/ugorji/go/codec"
	"gorm.io/gorm"
)

// TestMessageHandleIsIdempotentAndOwnerIsolated 验证对应测试场景。
func TestMessageHandleIsIdempotentAndOwnerIsolated(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Message{}, &model.ConversationTakeover{}))
	account := model.Account{OwnerID: 10, Label: "消息账号", Status: "ready", Enabled: true}
	require.NoError(t, db.Create(&account).Error)
	raw, err := os.ReadFile("../xianyu/protocol/testdata/sync_message.json")
	require.NoError(t, err)

	service := &Message{db: db, events: &BusinessEvent{db: db}}
	var workers sync.WaitGroup
	errorsCh := make(chan error, 8)
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			errorsCh <- service.Handle(t.Context(), account.ID, raw)
		}()
	}
	workers.Wait()
	close(errorsCh)
	for err := range errorsCh {
		require.NoError(t, err)
	}

	list, total, err := service.List(t.Context(), 10, xyRequest.MessageSearch{})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, list, 1)
	require.Equal(t, "message-demo.PNM", list[0].PNMID)
	other, total, err := service.List(t.Context(), 11, xyRequest.MessageSearch{})
	require.NoError(t, err)
	require.Zero(t, total)
	require.Empty(t, other)
}

// TestMessageListFiltersByExplicitFields 验证对应测试场景。
func TestMessageListFiltersByExplicitFields(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Message{}, &model.ConversationTakeover{}))
	account := model.Account{OwnerID: 10, Label: "消息账号", Status: "ready", Enabled: true}
	require.NoError(t, db.Create(&account).Error)
	for _, item := range []model.Message{
		{OwnerID: 10, AccountID: account.ID, PNMID: "in-1", Content: "价格是多少", SenderUserName: "买家甲", ExternalProductID: "item-1", MessageTime: 1000, Direction: "inbound", Source: "remote"},
		{OwnerID: 10, AccountID: account.ID, PNMID: "out-1", Content: "回复价格", SenderUserName: "卖家", ExternalProductID: "item-1", MessageTime: 1050, Direction: "outbound", Source: "remote"},
		{OwnerID: 10, AccountID: account.ID, PNMID: "in-2", Content: "价格是多少", SenderUserName: "买家乙", ExternalProductID: "item-2", MessageTime: 1100, Direction: "inbound", Source: "remote"},
	} {
		require.NoError(t, db.Create(&item).Error)
	}
	service := &Message{db: db, events: &BusinessEvent{db: db}}
	list, total, err := service.List(t.Context(), 10, xyRequest.MessageSearch{
		AccountID: account.ID, Direction: "inbound", ExternalProductID: "item-1",
		BuyerName: "甲", Content: "价格", StartTime: 1000, EndTime: 1000,
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, list, 1)
	require.Equal(t, "in-1", list[0].PNMID)
	_, _, err = service.List(t.Context(), 10, xyRequest.MessageSearch{Direction: "invalid"})
	require.ErrorContains(t, err, "消息方向无效")
	_, _, err = service.List(t.Context(), 10, xyRequest.MessageSearch{StartTime: 2, EndTime: 1})
	require.ErrorContains(t, err, "结束时间不能早于开始时间")
}

// TestMessageHandleMarksOwnRemoteMessageOutbound 验证对应测试场景。
func TestMessageHandleMarksOwnRemoteMessageOutbound(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Message{}, &model.ConversationTakeover{}))
	account := model.Account{OwnerID: 10, ExternalUserID: "seller-1", Label: "消息账号", Status: "ready", Enabled: true}
	require.NoError(t, db.Create(&account).Error)
	payload := map[string]any{"1": map[string]any{
		"2": "session@goofish", "3": "self-message", "5": int64(123),
		"10": map[string]any{"reminderContent": "我发出的消息", "senderUserId": "seller-1@goofish"},
	}}
	var packed []byte
	handle := new(codec.MsgpackHandle)
	require.NoError(t, codec.NewEncoderBytes(&packed, handle).Encode(payload))
	raw, err := json.Marshal(map[string]any{
		"lwp": protocol.SyncPush,
		"body": map[string]any{"syncPushPackage": map[string]any{"data": []any{
			map[string]any{"data": base64.RawStdEncoding.EncodeToString(packed)},
		}}},
	})
	require.NoError(t, err)

	service := &Message{db: db, events: &BusinessEvent{db: db}}
	require.NoError(t, service.Handle(t.Context(), account.ID, raw))
	var saved model.Message
	require.NoError(t, db.First(&saved).Error)
	require.Equal(t, "outbound", saved.Direction)
}

// TestManualReplyTakesOverAndRecordsConfirmedMessage 验证对应测试场景。
func TestManualReplyTakesOverAndRecordsConfirmedMessage(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Message{}, &model.ConversationTakeover{}))
	account := model.Account{OwnerID: 10, Label: "消息账号", Status: "ready", Enabled: true}
	require.NoError(t, db.Create(&account).Error)
	incoming := model.Message{
		OwnerID: 10, AccountID: account.ID, PNMID: "incoming-1", SessionID: "session-1@goofish",
		ContentType: 1, Content: "你好", SenderUserID: "buyer-1", ExternalProductID: "item-1",
		Direction: "inbound", Source: "remote",
	}
	require.NoError(t, db.Create(&incoming).Error)
	sender := new(fakeTextSender)
	service := &Message{db: db, events: &BusinessEvent{db: db}, sender: sender}

	require.Error(t, service.Reply(t.Context(), 11, xyRequest.SendMessage{MessageID: incoming.ID, Text: "您好"}))
	require.NoError(t, service.Reply(t.Context(), 10, xyRequest.SendMessage{MessageID: incoming.ID, Text: "您好"}))
	require.Equal(t, "session-1@goofish", sender.sessionID)
	require.Equal(t, "buyer-1", sender.toUserID)

	list, total, err := service.List(t.Context(), 10, xyRequest.MessageSearch{})
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Equal(t, "outbound", list[0].Direction)
	require.Equal(t, "manual", list[0].Source)
	require.True(t, list[0].TakeoverActive)
	require.NoError(t, service.UpdateTakeover(t.Context(), 10, xyRequest.UpdateTakeover{MessageID: incoming.ID, Active: false}))
	list, _, err = service.List(t.Context(), 10, xyRequest.MessageSearch{})
	require.NoError(t, err)
	require.False(t, list[0].TakeoverActive)
}

// fakeTextSender 是测试辅助类型。
type fakeTextSender struct {
	sessionID string // sessionID 是会话标识。
	toUserID  string // toUserID 是接收方标识。
}

// SendText 提供测试辅助能力。
func (s *fakeTextSender) SendText(_ context.Context, _, _ uint, sessionID, toUserID, _ string) error {
	s.sessionID = sessionID
	s.toUserID = toUserID
	return nil
}

// TestAutoReplyMergesMessagesAndSendsOnlyLatest 验证对应测试场景。
func TestAutoReplyMergesMessagesAndSendsOnlyLatest(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.Account{}, &model.Product{}, &model.ProductSetting{}, &model.ReplyRule{}, &model.ReplyContent{},
		&model.Message{}, &model.ConversationTakeover{}, &model.ReplyTask{},
	))
	account := model.Account{OwnerID: 10, ExternalUserID: "seller-1", Label: "消息账号", Status: "ready", Enabled: true}
	require.NoError(t, db.Create(&account).Error)
	product := model.Product{OwnerID: 10, AccountID: account.ID, ExternalProductID: "item-1", Title: "商品", Status: "on_sale"}
	require.NoError(t, db.Create(&product).Error)
	require.NoError(t, db.Create(&model.ProductSetting{ProductID: product.ID, AutoReplyEnabled: true, ReplyDelaySeconds: 1}).Error)
	rules := &ReplyRule{db: db}
	require.NoError(t, rules.Save(t.Context(), 10, xyRequest.SaveReplyRule{
		ProductID: product.ID, Type: model.ReplyRuleFuzzy, Keyword: "你好", Content: "自动回复内容", Enabled: true,
	}))
	sent := make(chan string, 2)
	service := &Message{db: db, events: &BusinessEvent{db: db}, sender: channelTextSender(sent), rules: rules, pending: make(map[string]PendingReply)}
	first := model.Message{OwnerID: 10, AccountID: account.ID, PNMID: "one", SessionID: "session-1", ContentType: 1, Content: "你好", SenderUserID: "buyer-1", ExternalProductID: "item-1", Direction: "inbound", Source: "remote"}
	second := first
	second.PNMID = "two"
	require.NoError(t, db.Create(&first).Error)
	require.NoError(t, service.scheduleAutoReply(t.Context(), account, first))
	require.NoError(t, db.Create(&second).Error)
	require.NoError(t, service.scheduleAutoReply(t.Context(), account, second))

	select {
	case text := <-sent:
		require.Equal(t, "自动回复内容", text)
	case <-time.After(2 * time.Second):
		t.Fatal("自动回复未发送")
	}
	select {
	case <-sent:
		t.Fatal("同一会话发送了重复回复")
	case <-time.After(100 * time.Millisecond):
	}
	var tasks []model.ReplyTask
	require.NoError(t, db.Order("id").Find(&tasks).Error)
	require.Len(t, tasks, 2)
	require.Equal(t, model.ReplyTaskCancelled, tasks[0].Status)
	require.Equal(t, model.ReplyTaskSent, tasks[1].Status)
}

// TestVirtualProductUsesDefaultFallbackReply 验证对应测试场景。
func TestVirtualProductUsesDefaultFallbackReply(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.Account{}, &model.Product{}, &model.ProductSetting{}, &model.ReplyRule{}, &model.ReplyContent{},
		&model.Message{}, &model.ConversationTakeover{}, &model.ReplyTask{},
	))
	account := model.Account{OwnerID: 10, ExternalUserID: "seller-1", Label: "消息账号", Status: "ready", Enabled: true}
	require.NoError(t, db.Create(&account).Error)
	product := model.Product{OwnerID: 10, AccountID: account.ID, ExternalProductID: "item-1", Title: "虚拟商品", Status: "on_sale"}
	require.NoError(t, db.Create(&product).Error)
	require.NoError(t, db.Create(&model.ProductSetting{ProductID: product.ID, VirtualProduct: true, AutoReplyEnabled: true}).Error)
	incoming := model.Message{
		OwnerID: 10, AccountID: account.ID, PNMID: "one", SessionID: "session-1", ContentType: 1,
		Content: "你好", SenderUserID: "buyer-1", ExternalProductID: "item-1", Direction: "inbound", Source: "remote",
	}
	require.NoError(t, db.Create(&incoming).Error)
	sent := make(chan string, 1)
	service := &Message{db: db, events: &BusinessEvent{db: db}, sender: channelTextSender(sent), rules: &ReplyRule{db: db}, pending: make(map[string]PendingReply)}

	require.NoError(t, service.scheduleAutoReply(t.Context(), account, incoming))
	select {
	case text := <-sent:
		require.Equal(t, virtualProductFallbackReply, text)
	case <-time.After(time.Second):
		t.Fatal("虚拟商品兜底回复未发送")
	}
}

// TestHandleHistoryRepliesOnlyToLatestRecentMessage 验证对应测试场景。
func TestHandleHistoryRepliesOnlyToLatestRecentMessage(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.Account{}, &model.Product{}, &model.ProductSetting{}, &model.ReplyRule{}, &model.ReplyContent{},
		&model.Message{}, &model.ConversationTakeover{}, &model.ReplyTask{},
	))
	account := model.Account{OwnerID: 10, ExternalUserID: "seller", Label: "账号", Status: "ready", Enabled: true}
	require.NoError(t, db.Create(&account).Error)
	product := model.Product{OwnerID: 10, AccountID: account.ID, ExternalProductID: "item-1", Title: "商品", Status: "on_sale"}
	require.NoError(t, db.Create(&product).Error)
	require.NoError(t, db.Create(&model.ProductSetting{ProductID: product.ID, AutoReplyEnabled: true}).Error)
	rules := &ReplyRule{db: db}
	require.NoError(t, rules.Save(t.Context(), 10, xyRequest.SaveReplyRule{
		ProductID: product.ID, Type: model.ReplyRuleFuzzy, Keyword: "价格是什么", Content: "十元", Enabled: true,
	}))
	sent := make(chan string, 2)
	service := &Message{db: db, events: &BusinessEvent{db: db}, sender: channelTextSender(sent), rules: rules, pending: make(map[string]PendingReply)}
	now := time.Now().UnixMilli()
	raw := historyResponse(t, []map[string]any{
		historyMessage(t, "old", "buyer", "价格是什么", "item-1", now-1000),
		historyMessage(t, "latest", "buyer", "价格是什么", "item-1", now),
	})

	require.NoError(t, service.HandleHistory(t.Context(), account.ID, "session@goofish", raw))
	select {
	case text := <-sent:
		require.Equal(t, "十元", text)
	case <-time.After(time.Second):
		t.Fatal("补偿回复未发送")
	}
	select {
	case <-sent:
		t.Fatal("历史补偿重复回复")
	case <-time.After(100 * time.Millisecond):
	}
}

// TestHandleHistoryMergesConfirmedOutboundEcho 验证对应测试场景。
func TestHandleHistoryMergesConfirmedOutboundEcho(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Message{}, &model.ConversationTakeover{}))
	account := model.Account{OwnerID: 10, ExternalUserID: "seller", Label: "账号", Status: "ready", Enabled: true}
	require.NoError(t, db.Create(&account).Error)
	now := time.Now().UnixMilli()
	require.NoError(t, db.Create(&model.Message{
		OwnerID: 10, AccountID: account.ID, PNMID: "local-1", SessionID: "session@goofish",
		ContentType: 1, Content: "已回复", MessageTime: now, Direction: "outbound", Source: "keyword",
	}).Error)
	raw := historyResponse(t, []map[string]any{historyMessage(t, "remote-1", "seller@goofish", "已回复", "item-1", now)})

	service := &Message{db: db, events: &BusinessEvent{db: db}}
	require.NoError(t, service.HandleHistory(t.Context(), account.ID, "session@goofish", raw))
	var saved []model.Message
	require.NoError(t, db.Find(&saved).Error)
	require.Len(t, saved, 1)
	require.Equal(t, "remote-1", saved[0].PNMID)
	require.Equal(t, "remote", saved[0].Source)
}

// historyMessage 提供测试辅助能力。
func historyMessage(t *testing.T, id, sender, content, itemID string, timestamp int64) map[string]any {
	t.Helper()
	contentJSON, err := json.Marshal(map[string]any{"contentType": 1, "text": map[string]string{"text": content}})
	require.NoError(t, err)
	encoded := base64.StdEncoding.EncodeToString(contentJSON)
	return map[string]any{"message": map[string]any{
		"messageId": id, "createAt": timestamp,
		"extension": map[string]any{"senderUserId": sender, "reminderUrl": "https://www.goofish.com/?itemId=" + itemID},
		"content":   map[string]any{"custom": map[string]any{"data": encoded}},
	}}
}

// historyResponse 提供测试辅助能力。
func historyResponse(t *testing.T, messages []map[string]any) []byte {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"body": map[string]any{"userMessageModels": messages}})
	require.NoError(t, err)
	return raw
}

// channelTextSender 是测试辅助类型。
type channelTextSender chan string

// SendText 提供测试辅助能力。
func (s channelTextSender) SendText(_ context.Context, _, _ uint, _, _, text string) error {
	s <- text
	return nil
}
