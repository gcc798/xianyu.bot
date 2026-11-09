package service

import (
	"context"
	"errors"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/response"
	"gorm.io/gorm"
)

// Status 提供数据库健康检查和用户业务概览。
type Status struct {
	db *gorm.DB // db 用于健康检查和聚合统计。
}

// NewStatus 创建状态查询服务。
func NewStatus(db *gorm.DB) *Status {
	return &Status{db: db}
}

// Get 检查数据库并汇总 ownerID 名下的关键业务数量。
func (s *Status) Get(ctx context.Context, ownerID uint) (xyResponse.Status, error) {
	if s.db == nil {
		return xyResponse.Status{}, errors.New("database is not initialized")
	}
	if err := s.db.WithContext(ctx).Exec("SELECT 1").Error; err != nil {
		return xyResponse.Status{}, err
	}
	result := xyResponse.Status{Database: "postgresql"}
	db := s.db.WithContext(ctx)
	for _, count := range []struct {
		model any    // model 指定 GORM 统计表。
		query string // query 是带占位符的归属与状态条件。
		args  []any  // args 是 query 对应参数。
		value *int64 // value 接收统计结果。
	}{
		{&model.Account{}, "owner_id = ?", []any{ownerID}, &result.Accounts},
		{&model.Account{}, "owner_id = ? AND status = ? AND enabled = ?", []any{ownerID, "ready", true}, &result.ReadyAccounts},
		{&model.Product{}, "owner_id = ? AND status = ?", []any{ownerID, "on_sale"}, &result.OnSaleProducts},
		{&model.Message{}, "owner_id = ?", []any{ownerID}, &result.Messages},
		{&model.Order{}, "owner_id = ? AND status = ?", []any{ownerID, model.OrderStatusPendingDelivery}, &result.PendingOrders},
		{&model.DeliveryTask{}, "owner_id = ? AND status IN ?", []any{ownerID, []string{model.DeliveryTaskFailed, model.DeliveryTaskUnknown}}, &result.FailedDeliveries},
	} {
		if err := db.Model(count.model).Where(count.query, count.args...).Count(count.value).Error; err != nil {
			return xyResponse.Status{}, err
		}
	}
	return result, nil
}
