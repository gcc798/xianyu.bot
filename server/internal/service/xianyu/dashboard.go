package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model/xianyu"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/xianyu/request"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/xianyu/response"
	"gorm.io/gorm"
)

const (
	dashboardHour = int64(time.Hour / time.Millisecond)
	dashboardDay  = int64(24 * time.Hour / time.Millisecond)
)

func (s *status) Overview(ctx context.Context, ownerID uint, input xyRequest.DashboardOverview) (xyResponse.DashboardOverview, error) {
	if ownerID == 0 {
		return xyResponse.DashboardOverview{}, errors.New("用户未登录")
	}
	if input.StartTime <= 0 || input.EndTime <= input.StartTime {
		return xyResponse.DashboardOverview{}, errors.New("日期范围无效")
	}
	if input.EndTime-input.StartTime > int64(90*24*time.Hour/time.Millisecond) {
		return xyResponse.DashboardOverview{}, errors.New("工作台单次最多查询 90 天")
	}
	if s.db == nil {
		return xyResponse.DashboardOverview{}, errors.New("数据库未初始化")
	}

	bucketMillis := dashboardDay
	if input.EndTime-input.StartTime <= 48*dashboardHour {
		bucketMillis = dashboardHour
	}
	overview := xyResponse.DashboardOverview{
		StartTime: input.StartTime, EndTime: input.EndTime, BucketMillis: bucketMillis,
		MessageTrend: makeMessageTrend(input.StartTime, input.EndTime, bucketMillis),
		OrderTrend:   makeOrderTrend(input.StartTime, input.EndTime, bucketMillis),
		RecentEvents: make([]xyResponse.BusinessEvent, 0),
	}
	db := s.db.WithContext(ctx)
	if err := loadDashboardMetrics(db, ownerID, input, &overview); err != nil {
		return xyResponse.DashboardOverview{}, err
	}
	if err := loadMessageTrend(db, ownerID, input, &overview); err != nil {
		return xyResponse.DashboardOverview{}, err
	}
	if err := loadOrderTrend(db, ownerID, input, &overview); err != nil {
		return xyResponse.DashboardOverview{}, err
	}
	if err := loadRecentEvents(db, ownerID, input, &overview.RecentEvents); err != nil {
		return xyResponse.DashboardOverview{}, err
	}
	return overview, nil
}

func loadDashboardMetrics(db *gorm.DB, ownerID uint, input xyRequest.DashboardOverview, overview *xyResponse.DashboardOverview) error {
	start, end := time.UnixMilli(input.StartTime), time.UnixMilli(input.EndTime)
	counts := []struct {
		model         any
		accountColumn string
		query         string
		args          []any
		value         *int64
	}{
		{&model.Account{}, "id", "owner_id = ?", []any{ownerID}, &overview.Metrics.Accounts},
		{&model.Account{}, "id", "owner_id = ? AND status = ? AND enabled = ?", []any{ownerID, "ready", true}, &overview.Metrics.ReadyAccounts},
		{&model.Product{}, "account_id", "owner_id = ? AND status = ?", []any{ownerID, "on_sale"}, &overview.Metrics.OnSaleProducts},
		{&model.Message{}, "account_id", "owner_id = ? AND direction = ? AND message_time >= ? AND message_time < ?", []any{ownerID, "inbound", input.StartTime, input.EndTime}, &overview.Metrics.InboundMessages},
		{&model.Message{}, "account_id", "owner_id = ? AND direction = ? AND message_time >= ? AND message_time < ?", []any{ownerID, "outbound", input.StartTime, input.EndTime}, &overview.Metrics.OutboundMessages},
		{&model.ReplyTask{}, "account_id", "owner_id = ? AND status = ? AND completed_at >= ? AND completed_at < ?", []any{ownerID, model.ReplyTaskSent, start, end}, &overview.Automation.ReplySuccess},
		{&model.ReplyTask{}, "account_id", "owner_id = ? AND status IN ? AND completed_at >= ? AND completed_at < ?", []any{ownerID, []string{model.ReplyTaskFailed, model.ReplyTaskUnknown}, start, end}, &overview.Automation.ReplyFailed},
		{&model.Order{}, "account_id", "owner_id = ? AND created_at >= ? AND created_at < ?", []any{ownerID, start, end}, &overview.Metrics.NewOrders},
		{&model.Order{}, "account_id", "owner_id = ? AND status = ?", []any{ownerID, model.OrderStatusPendingDelivery}, &overview.Metrics.PendingOrders},
		{&model.DeliveryTask{}, "account_id", "owner_id = ? AND status IN ? AND updated_at >= ? AND updated_at < ?", []any{ownerID, []string{model.DeliveryTaskConfirmed, model.DeliveryTaskSentUnconfirmed}, start, end}, &overview.Automation.DeliverySuccess},
		{&model.DeliveryTask{}, "account_id", "owner_id = ? AND status IN ? AND updated_at >= ? AND updated_at < ?", []any{ownerID, []string{model.DeliveryTaskFailed, model.DeliveryTaskUnknown}, start, end}, &overview.Automation.DeliveryFailed},
		{&model.RelistTask{}, "account_id", "owner_id = ? AND status = ? AND updated_at >= ? AND updated_at < ?", []any{ownerID, model.RelistTaskSucceeded, start, end}, &overview.Automation.RelistSuccess},
		{&model.RelistTask{}, "account_id", "owner_id = ? AND status IN ? AND updated_at >= ? AND updated_at < ?", []any{ownerID, []string{model.RelistTaskFailed, model.RelistTaskUnknown}, start, end}, &overview.Automation.RelistFailed},
	}
	for _, count := range counts {
		query := db.Model(count.model).Where(count.query, count.args...)
		if input.AccountID != 0 {
			query = query.Where(count.accountColumn+" = ?", input.AccountID)
		}
		if err := query.Count(count.value).Error; err != nil {
			return fmt.Errorf("统计工作台指标: %w", err)
		}
	}
	overview.Metrics.AutoReplies = overview.Automation.ReplySuccess
	overview.Metrics.FailedAutomationTasks = overview.Automation.ReplyFailed + overview.Automation.DeliveryFailed + overview.Automation.RelistFailed
	if overview.Metrics.InboundMessages > 0 {
		overview.Metrics.AutoReplyRate = float64(overview.Metrics.AutoReplies) * 100 / float64(overview.Metrics.InboundMessages)
	}
	otherOutbound := max(overview.Metrics.OutboundMessages-overview.Metrics.AutoReplies, 0)
	overview.MessageSources = []xyResponse.DashboardNameValue{
		{Name: "买家发来", Value: overview.Metrics.InboundMessages},
		{Name: "自动回复", Value: overview.Metrics.AutoReplies},
		{Name: "其他发出", Value: otherOutbound},
	}
	return nil
}

func loadMessageTrend(db *gorm.DB, ownerID uint, input xyRequest.DashboardOverview, overview *xyResponse.DashboardOverview) error {
	type messageBucket struct {
		BucketIndex int64
		Inbound     int64
		Outbound    int64
	}
	var messageBuckets []messageBucket
	query := db.Model(&model.Message{}).
		Select("((message_time - ?) / ?) AS bucket_index, SUM(CASE WHEN direction = 'inbound' THEN 1 ELSE 0 END) AS inbound, SUM(CASE WHEN direction = 'outbound' THEN 1 ELSE 0 END) AS outbound", input.StartTime, overview.BucketMillis).
		Where("owner_id = ? AND message_time >= ? AND message_time < ?", ownerID, input.StartTime, input.EndTime)
	if input.AccountID != 0 {
		query = query.Where("account_id = ?", input.AccountID)
	}
	if err := query.Group("bucket_index").Scan(&messageBuckets).Error; err != nil {
		return fmt.Errorf("统计消息趋势: %w", err)
	}
	for _, bucket := range messageBuckets {
		if bucket.BucketIndex >= 0 && bucket.BucketIndex < int64(len(overview.MessageTrend)) {
			overview.MessageTrend[bucket.BucketIndex].Inbound = bucket.Inbound
			overview.MessageTrend[bucket.BucketIndex].Outbound = bucket.Outbound
		}
	}
	var autoReplies []struct{ CompletedAt time.Time }
	autoQuery := db.Model(&model.ReplyTask{}).Select("completed_at").
		Where("owner_id = ? AND status = ? AND completed_at >= ? AND completed_at < ?", ownerID, model.ReplyTaskSent, time.UnixMilli(input.StartTime), time.UnixMilli(input.EndTime))
	if input.AccountID != 0 {
		autoQuery = autoQuery.Where("account_id = ?", input.AccountID)
	}
	if err := autoQuery.Scan(&autoReplies).Error; err != nil {
		return fmt.Errorf("统计自动回复趋势: %w", err)
	}
	for _, reply := range autoReplies {
		index := (reply.CompletedAt.UnixMilli() - input.StartTime) / overview.BucketMillis
		if index >= 0 && index < int64(len(overview.MessageTrend)) {
			overview.MessageTrend[index].AutoReplies++
		}
	}
	return nil
}

func loadOrderTrend(db *gorm.DB, ownerID uint, input xyRequest.DashboardOverview, overview *xyResponse.DashboardOverview) error {
	start, end := time.UnixMilli(input.StartTime), time.UnixMilli(input.EndTime)
	var orders []struct{ CreatedAt time.Time }
	orderQuery := db.Model(&model.Order{}).Select("created_at").Where("owner_id = ? AND created_at >= ? AND created_at < ?", ownerID, start, end)
	if input.AccountID != 0 {
		orderQuery = orderQuery.Where("account_id = ?", input.AccountID)
	}
	if err := orderQuery.Scan(&orders).Error; err != nil {
		return fmt.Errorf("统计订单趋势: %w", err)
	}
	for _, order := range orders {
		if index := (order.CreatedAt.UnixMilli() - input.StartTime) / overview.BucketMillis; index >= 0 && index < int64(len(overview.OrderTrend)) {
			overview.OrderTrend[index].Orders++
		}
	}
	var deliveries []struct {
		UpdatedAt time.Time
		Status    string
	}
	deliveryQuery := db.Model(&model.DeliveryTask{}).Select("updated_at", "status").
		Where("owner_id = ? AND updated_at >= ? AND updated_at < ?", ownerID, start, end)
	if input.AccountID != 0 {
		deliveryQuery = deliveryQuery.Where("account_id = ?", input.AccountID)
	}
	if err := deliveryQuery.Scan(&deliveries).Error; err != nil {
		return fmt.Errorf("统计发货趋势: %w", err)
	}
	for _, delivery := range deliveries {
		index := (delivery.UpdatedAt.UnixMilli() - input.StartTime) / overview.BucketMillis
		if index < 0 || index >= int64(len(overview.OrderTrend)) {
			continue
		}
		switch delivery.Status {
		case model.DeliveryTaskConfirmed, model.DeliveryTaskSentUnconfirmed:
			overview.OrderTrend[index].Delivered++
		case model.DeliveryTaskFailed, model.DeliveryTaskUnknown:
			overview.OrderTrend[index].Failed++
		}
	}
	return nil
}

func loadRecentEvents(db *gorm.DB, ownerID uint, input xyRequest.DashboardOverview, events *[]xyResponse.BusinessEvent) error {
	query := db.Table("xy_business_events AS e").
		Select(`e.id, e.created_at, e.account_id, COALESCE(a.label, '') AS account_label,
			e.product_id, COALESCE(p.title, '') AS product_title, e.order_id, COALESCE(o.external_order_id, '') AS external_order_id,
			e.type, e.status, e.summary, e.error`).
		Joins("LEFT JOIN xy_accounts AS a ON a.id = e.account_id").
		Joins("LEFT JOIN xy_products AS p ON p.id = e.product_id").
		Joins("LEFT JOIN xy_orders AS o ON o.id = e.order_id").
		Where("e.owner_id = ? AND e.deleted_at IS NULL AND e.created_at >= ? AND e.created_at < ?", ownerID, time.UnixMilli(input.StartTime), time.UnixMilli(input.EndTime))
	if input.AccountID != 0 {
		query = query.Where("e.account_id = ?", input.AccountID)
	}
	if err := query.Order("e.id DESC").Limit(8).Scan(events).Error; err != nil {
		return fmt.Errorf("查询最近业务事件: %w", err)
	}
	return nil
}

func makeMessageTrend(start, end, bucket int64) []xyResponse.DashboardMessagePoint {
	count := (end - start + bucket - 1) / bucket
	points := make([]xyResponse.DashboardMessagePoint, count)
	for index := range count {
		points[index].Timestamp = start + int64(index)*bucket
	}
	return points
}

func makeOrderTrend(start, end, bucket int64) []xyResponse.DashboardOrderPoint {
	count := (end - start + bucket - 1) / bucket
	points := make([]xyResponse.DashboardOrderPoint, count)
	for index := range count {
		points[index].Timestamp = start + int64(index)*bucket
	}
	return points
}
