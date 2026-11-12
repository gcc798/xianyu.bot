package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// blockingDeliverySender 是测试辅助类型。
type blockingDeliverySender struct {
	payload chan string   // payload 记录发货内容。
	release chan struct{} // release 控制模拟发货完成。
}

// Consign 提供测试辅助能力。
func (s *blockingDeliverySender) Consign(ctx context.Context, _ uint, _ model.Order, text string) error {
	s.payload <- text
	select {
	case <-s.release:
		return nil
	case <-ctx.Done():
		return context.Cause(ctx)
	}
}

// TestDeliveryReservesEncryptedCardOnce 验证对应测试场景。
func TestDeliveryReservesEncryptedCardOnce(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)
	t.Setenv(credentialKeyEnvironment, base64.StdEncoding.EncodeToString(key))
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "delivery.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Product{}, &model.ProductSetting{}, &model.DeliveryRule{}, &model.DeliveryResource{}, &model.Order{}, &model.DeliveryTask{}))

	account := model.Account{OwnerID: 10, Label: "发货账号", Status: "ready", Enabled: true}
	require.NoError(t, db.Create(&account).Error)
	product := model.Product{OwnerID: 10, AccountID: account.ID, ExternalProductID: "item-1", Title: "商品", Status: "on_sale"}
	require.NoError(t, db.Create(&product).Error)
	require.NoError(t, db.Create(&model.ProductSetting{ProductID: product.ID, AutoDeliveryEnabled: true}).Error)
	rule := model.DeliveryRule{OwnerID: 10, ProductID: product.ID, Mode: model.DeliveryModeCard, Template: "卡密：{card}", Enabled: true}
	require.NoError(t, db.Create(&rule).Error)
	sender := &blockingDeliverySender{payload: make(chan string, 1), release: make(chan struct{})}
	service := &Delivery{db: db, events: &BusinessEvent{db: db}, consigner: sender, rules: &DeliveryRule{db: db}, workers: make(map[uint]context.CancelFunc)}
	count, err := service.ImportResources(t.Context(), 10, xyRequest.ImportDeliveryResources{RuleID: rule.ID, Values: []string{"SECRET-CARD"}})
	require.NoError(t, err)
	require.Equal(t, 1, count)
	_, err = service.ImportResources(t.Context(), 11, xyRequest.ImportDeliveryResources{RuleID: rule.ID, Values: []string{"FOREIGN-SECRET"}})
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	var resourceCount int64
	require.NoError(t, db.Model(&model.DeliveryResource{}).Count(&resourceCount).Error)
	require.EqualValues(t, 1, resourceCount)
	_, err = service.Inventory(t.Context(), 11, rule.ID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	first := model.Order{OwnerID: 10, AccountID: account.ID, ProductID: product.ID, ExternalOrderID: "order-1", BuyerUserID: "buyer", SessionID: "session", Status: model.OrderStatusPendingDelivery}
	require.NoError(t, db.Create(&first).Error)
	require.NoError(t, service.Plan(t.Context(), first))
	select {
	case payload := <-sender.payload:
		require.Equal(t, "卡密：SECRET-CARD", payload)
	case <-time.After(time.Second):
		t.Fatal("发货任务未启动")
	}
	var task model.DeliveryTask
	require.NoError(t, db.First(&task, "order_id = ?", first.ID).Error)
	require.NotContains(t, task.PayloadCiphertext, "SECRET-CARD")
	var resource model.DeliveryResource
	require.NoError(t, db.First(&resource).Error)
	require.Equal(t, model.DeliveryResourceReserved, resource.Status)
	require.False(t, strings.Contains(resource.ValueCiphertext, "SECRET-CARD"))

	second := model.Order{OwnerID: 10, AccountID: account.ID, ProductID: product.ID, ExternalOrderID: "order-2", BuyerUserID: "buyer", SessionID: "session", Status: model.OrderStatusPendingDelivery}
	require.NoError(t, db.Create(&second).Error)
	require.NoError(t, service.Plan(t.Context(), second))
	var secondTask model.DeliveryTask
	require.NoError(t, db.First(&secondTask, "order_id = ?", second.ID).Error)
	require.Equal(t, model.DeliveryTaskFailed, secondTask.Status)
	require.Equal(t, "卡密库存不足", secondTask.LastError)
	list, total, err := service.ListTasks(t.Context(), 10, xyRequest.DeliveryTaskSearch{})
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, list, 2)
	list, total, err = service.ListTasks(t.Context(), 10, xyRequest.DeliveryTaskSearch{ProductID: product.ID})
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, list, 2)
	otherOwnerTasks, otherOwnerTotal, err := service.ListTasks(t.Context(), 11, xyRequest.DeliveryTaskSearch{})
	require.NoError(t, err)
	require.Zero(t, otherOwnerTotal)
	require.Empty(t, otherOwnerTasks)

	close(sender.release)
	require.Eventually(t, func() bool {
		return db.First(&task, task.ID).Error == nil && task.Status == model.DeliveryTaskConfirmed
	}, time.Second, 10*time.Millisecond)
	require.NoError(t, db.First(&resource, resource.ID).Error)
	require.Equal(t, model.DeliveryResourceConsumed, resource.Status)
}
