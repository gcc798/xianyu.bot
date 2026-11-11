package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// fakeMTopCaller 是测试辅助类型。
type fakeMTopCaller struct {
	data json.RawMessage // data 是模拟响应。
	err  error           // err 是模拟错误。
}

// recordedMTopCall 是测试辅助类型。
type recordedMTopCall struct {
	api    string // api 是调用接口名。
	data   any    // data 是调用参数。
	seller bool   // seller 标记卖家接口调用。
}

// recordingMTopCaller 是测试辅助类型。
type recordingMTopCaller struct {
	calls []recordedMTopCall // calls 记录 MTop 调用。
}

// Call 提供测试辅助能力。
func (f *recordingMTopCaller) Call(_ context.Context, _, _ uint, api string, data any, seller bool) (json.RawMessage, error) {
	f.calls = append(f.calls, recordedMTopCall{api: api, data: data, seller: seller})
	return json.RawMessage(`{}`), nil
}

// recordingTextSender 是测试辅助类型。
type recordingTextSender struct {
	texts []string // texts 记录已发送文本。
}

// SendText 提供测试辅助能力。
func (s *recordingTextSender) SendText(_ context.Context, _, _ uint, sessionID, toUserID, text string) error {
	if sessionID == "" || toUserID == "" {
		return errors.New("缺少会话信息")
	}
	s.texts = append(s.texts, text)
	return nil
}

// Call 提供测试辅助能力。
func (f fakeMTopCaller) Call(context.Context, uint, uint, string, any, bool) (json.RawMessage, error) {
	return f.data, f.err
}

// TestOrderEventsAreIdempotentAndOwnerScoped 验证对应测试场景。
func TestOrderEventsAreIdempotentAndOwnerScoped(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Product{}, &model.Order{}, &model.OrderEvent{}))
	account := model.Account{OwnerID: 10, Label: "订单账号", Status: "ready", Enabled: true}
	require.NoError(t, db.Create(&account).Error)
	product := model.Product{OwnerID: 10, AccountID: account.ID, ExternalProductID: "item-1", Title: "商品", Status: "on_sale"}
	require.NoError(t, db.Create(&product).Error)
	message := model.Message{
		AccountID: account.ID, PNMID: "event-1", SessionID: "session-1", ContentType: 26,
		SenderUserID: "buyer-1", SenderUserName: "买家", ExternalProductID: "item-1", ExternalOrderID: "order-1",
		MessageTime: 100,
	}
	service := &Order{db: db, events: &BusinessEvent{db: db}}
	_, err = service.UpsertFromMessage(t.Context(), account, message)
	require.NoError(t, err)
	_, err = service.UpsertFromMessage(t.Context(), account, message)
	require.NoError(t, err)

	list, total, err := service.List(t.Context(), 10, xyRequest.OrderSearch{})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, model.OrderStatusPendingDelivery, list[0].Status)
	require.Equal(t, "商品", list[0].ProductTitle)
	list, total, err = service.List(t.Context(), 10, xyRequest.OrderSearch{ProductID: product.ID})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, list, 1)
	other, total, err := service.List(t.Context(), 11, xyRequest.OrderSearch{})
	require.NoError(t, err)
	require.Zero(t, total)
	require.Empty(t, other)
	var events int64
	require.NoError(t, db.Model(&model.OrderEvent{}).Count(&events).Error)
	require.EqualValues(t, 1, events)
}

// TestOrderGetIsOwnerScoped 验证订单详情仅对所属用户可见。
func TestOrderGetIsOwnerScoped(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Order{}))
	account := model.Account{OwnerID: 10, Label: "订单账号", Status: "ready", Enabled: true}
	require.NoError(t, db.Create(&account).Error)
	item := model.Order{OwnerID: 10, AccountID: account.ID, ExternalOrderID: "order-1", ProductTitle: "测试商品", Status: model.OrderStatusPendingDelivery}
	require.NoError(t, db.Create(&item).Error)

	service := &Order{db: db}
	detail, err := service.Get(t.Context(), 10, item.ID)
	require.NoError(t, err)
	require.Equal(t, "order-1", detail.ExternalOrderID)
	_, err = service.Get(t.Context(), 11, item.ID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// TestSyncPendingOrdersUpsertsRemoteData 验证对应测试场景。
func TestSyncPendingOrdersUpsertsRemoteData(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Product{}, &model.Order{}))
	account := model.Account{OwnerID: 10, Label: "订单账号", Status: "ready", Enabled: true}
	require.NoError(t, db.Create(&account).Error)
	product := model.Product{OwnerID: 10, AccountID: account.ID, ExternalProductID: "item-1", Title: "本地标题"}
	require.NoError(t, db.Create(&product).Error)
	remote := json.RawMessage(`{"module":{"items":[{"commonData":{"orderId":"order-1","itemId":"item-1","orderStatus":"待发货"},"buyerInfoVO":{"userNick":"买家","userId":"buyer-1"},"itemVO":{"title":"远端标题"},"priceVO":{"totalPrice":"12.34"}}]}}`)
	service := &Order{db: db, events: &BusinessEvent{db: db}, mtop: fakeMTopCaller{data: remote}}
	count, err := service.SyncPending(t.Context(), 10, account.ID)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	count, err = service.SyncPending(t.Context(), 10, account.ID)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	var orders []model.Order
	require.NoError(t, db.Find(&orders).Error)
	require.Len(t, orders, 1)
	require.Equal(t, product.ID, orders[0].ProductID)
	require.Equal(t, int64(1234), orders[0].AmountCent)
	require.Equal(t, model.OrderStatusPendingDelivery, orders[0].Status)
}

// TestConsignSendsMessageThenConfirmsPlatformDelivery 验证对应测试场景。
func TestConsignSendsMessageThenConfirmsPlatformDelivery(t *testing.T) {
	mtop := new(recordingMTopCaller)
	sender := new(recordingTextSender)
	service := &Order{mtop: mtop, sender: sender}
	item := model.Order{AccountID: 2, ExternalOrderID: "order-1", SessionID: "session-1", BuyerUserID: "buyer-1"}

	require.NoError(t, service.Consign(t.Context(), 10, item, "网盘：https://example.test/share"))
	require.Equal(t, []string{"网盘：https://example.test/share"}, sender.texts)
	require.Len(t, mtop.calls, 1)
	require.Equal(t, "mtop.taobao.idle.logistic.consign.dummy", mtop.calls[0].api)
	require.False(t, mtop.calls[0].seller)
	require.Equal(t, map[string]any{
		"orderId": "order-1", "tradeText": "网盘：https://example.test/share", "picList": []any{}, "newUnconsign": true,
	}, mtop.calls[0].data)
}
