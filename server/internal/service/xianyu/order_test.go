package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/gcc798/xianyu.bot/internal/model/xianyu"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/xianyu/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type fakeMTopCaller struct {
	data json.RawMessage
	err  error
}

type recordedMTopCall struct {
	api    string
	data   any
	seller bool
}

type recordingMTopCaller struct{ calls []recordedMTopCall }

func (f *recordingMTopCaller) Call(_ context.Context, _, _ uint, api string, data any, seller bool) (json.RawMessage, error) {
	f.calls = append(f.calls, recordedMTopCall{api: api, data: data, seller: seller})
	return json.RawMessage(`{}`), nil
}

type recordingTextSender struct{ texts []string }

func (s *recordingTextSender) SendText(_ context.Context, _, _ uint, sessionID, toUserID, text string) error {
	if sessionID == "" || toUserID == "" {
		return errors.New("缺少会话信息")
	}
	s.texts = append(s.texts, text)
	return nil
}

func (f fakeMTopCaller) Call(context.Context, uint, uint, string, any, bool) (json.RawMessage, error) {
	return f.data, f.err
}

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
	service := &order{db: db, events: &businessEvent{db: db}}
	_, err = service.UpsertFromMessage(t.Context(), account, message)
	require.NoError(t, err)
	_, err = service.UpsertFromMessage(t.Context(), account, message)
	require.NoError(t, err)

	list, total, err := service.List(t.Context(), 10, xyRequest.OrderSearch{})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, model.OrderStatusPendingDelivery, list[0].Status)
	require.Equal(t, "商品", list[0].ProductTitle)
	other, total, err := service.List(t.Context(), 11, xyRequest.OrderSearch{})
	require.NoError(t, err)
	require.Zero(t, total)
	require.Empty(t, other)
	var events int64
	require.NoError(t, db.Model(&model.OrderEvent{}).Count(&events).Error)
	require.EqualValues(t, 1, events)
}

func TestSyncPendingOrdersUpsertsRemoteData(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Product{}, &model.Order{}))
	account := model.Account{OwnerID: 10, Label: "订单账号", Status: "ready", Enabled: true}
	require.NoError(t, db.Create(&account).Error)
	product := model.Product{OwnerID: 10, AccountID: account.ID, ExternalProductID: "item-1", Title: "本地标题"}
	require.NoError(t, db.Create(&product).Error)
	remote := json.RawMessage(`{"module":{"items":[{"commonData":{"orderId":"order-1","itemId":"item-1","orderStatus":"待发货"},"buyerInfoVO":{"userNick":"买家","userId":"buyer-1"},"itemVO":{"title":"远端标题"},"priceVO":{"totalPrice":"12.34"}}]}}`)
	service := &order{db: db, events: &businessEvent{db: db}, mtop: fakeMTopCaller{data: remote}}
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

func TestConsignSendsMessageThenConfirmsPlatformDelivery(t *testing.T) {
	mtop := new(recordingMTopCaller)
	sender := new(recordingTextSender)
	service := &order{mtop: mtop, sender: sender}
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
