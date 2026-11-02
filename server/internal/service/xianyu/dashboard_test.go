package service

import (
	"testing"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model/common"
	"github.com/gcc798/xianyu.bot/internal/model/xianyu"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/xianyu/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestDashboardOverviewIsOwnerAndAccountScoped(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.Account{}, &model.Product{}, &model.Message{}, &model.ReplyTask{},
		&model.Order{}, &model.DeliveryTask{}, &model.RelistTask{}, &model.BusinessEvent{},
	))

	start := time.Date(2026, time.October, 28, 0, 0, 0, 0, time.Local)
	end := start.Add(24 * time.Hour)
	account := model.Account{OwnerID: 10, Label: "主账号", Status: "ready", Enabled: true}
	otherAccount := model.Account{OwnerID: 10, Label: "次账号", Status: "ready", Enabled: true}
	require.NoError(t, db.Create(&account).Error)
	require.NoError(t, db.Create(&otherAccount).Error)
	require.NoError(t, db.Create(&model.Product{OwnerID: 10, AccountID: account.ID, ExternalProductID: "p-1", Title: "商品", Status: "on_sale"}).Error)

	for index, direction := range []string{"inbound", "inbound", "outbound"} {
		require.NoError(t, db.Create(&model.Message{
			OwnerID: 10, AccountID: account.ID, PNMID: direction + time.Duration(index).String(),
			Direction: direction, MessageTime: start.Add(time.Duration(index+1) * time.Hour).UnixMilli(),
		}).Error)
	}
	require.NoError(t, db.Create(&model.Message{OwnerID: 10, AccountID: otherAccount.ID, PNMID: "other-account", Direction: "inbound", MessageTime: start.Add(time.Hour).UnixMilli()}).Error)
	require.NoError(t, db.Create(&model.Message{OwnerID: 11, AccountID: 999, PNMID: "other-owner", Direction: "inbound", MessageTime: start.Add(time.Hour).UnixMilli()}).Error)

	completedAt := start.Add(2 * time.Hour)
	require.NoError(t, db.Create(&model.ReplyTask{
		OwnerID: 10, AccountID: account.ID, MessageID: 1, ExternalProductID: "p-1", SessionID: "s-1",
		RecipientID: "buyer", ReplyText: "回复", Status: model.ReplyTaskSent, ScheduledAt: completedAt, CompletedAt: &completedAt,
	}).Error)
	order := model.Order{
		Timestamps: common.Timestamps{CreatedAt: start.Add(3 * time.Hour)}, OwnerID: 10, AccountID: account.ID,
		ExternalOrderID: "o-1", Status: model.OrderStatusPendingDelivery,
	}
	require.NoError(t, db.Create(&order).Error)
	require.NoError(t, db.Create(&model.DeliveryTask{
		Timestamps: common.Timestamps{UpdatedAt: start.Add(4 * time.Hour)}, OwnerID: 10, AccountID: account.ID,
		OrderID: order.ID, PayloadCiphertext: "payload", Status: model.DeliveryTaskFailed, NextAttemptAt: end,
	}).Error)
	require.NoError(t, db.Create(&model.BusinessEvent{
		Timestamps: common.Timestamps{CreatedAt: start.Add(5 * time.Hour)}, OwnerID: 10, AccountID: account.ID,
		Type: "delivery", Status: "failed", Summary: "自动发货失败",
	}).Error)

	overview, err := (&status{db: db}).Overview(t.Context(), 10, xyRequest.DashboardOverview{
		AccountID: account.ID, StartTime: start.UnixMilli(), EndTime: end.UnixMilli(),
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, overview.Metrics.Accounts)
	require.EqualValues(t, 2, overview.Metrics.InboundMessages)
	require.EqualValues(t, 1, overview.Metrics.OutboundMessages)
	require.EqualValues(t, 1, overview.Metrics.AutoReplies)
	require.Equal(t, 50.0, overview.Metrics.AutoReplyRate)
	require.EqualValues(t, 1, overview.Metrics.NewOrders)
	require.EqualValues(t, 1, overview.Metrics.PendingOrders)
	require.EqualValues(t, 1, overview.Metrics.FailedAutomationTasks)
	require.Len(t, overview.MessageTrend, 24)
	require.EqualValues(t, 1, overview.MessageTrend[1].Inbound)
	require.EqualValues(t, 1, overview.MessageTrend[2].AutoReplies)
	require.EqualValues(t, 1, overview.OrderTrend[3].Orders)
	require.EqualValues(t, 1, overview.OrderTrend[4].Failed)
	require.Len(t, overview.RecentEvents, 1)
	require.Equal(t, "自动发货失败", overview.RecentEvents[0].Summary)
}

func TestDashboardOverviewRejectsMoreThanNinetyDays(t *testing.T) {
	start := time.Now()
	_, err := new(status).Overview(t.Context(), 1, xyRequest.DashboardOverview{
		StartTime: start.UnixMilli(), EndTime: start.Add(91 * 24 * time.Hour).UnixMilli(),
	})
	require.ErrorContains(t, err, "最多查询 90 天")
}
