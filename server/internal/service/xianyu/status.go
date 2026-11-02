package service

import (
	"context"
	"errors"

	"github.com/gcc798/go.xianyu.assistant/server/internal/infra"
	"github.com/gcc798/go.xianyu.assistant/server/internal/model/xianyu"
	xyResponse "github.com/gcc798/go.xianyu.assistant/server/internal/model/xianyu/response"
)

type status struct{}

func (s *status) Get(ctx context.Context, ownerID uint) (xyResponse.Status, error) {
	if infra.DB == nil {
		return xyResponse.Status{}, errors.New("database is not initialized")
	}
	if err := infra.DB.WithContext(ctx).Exec("SELECT 1").Error; err != nil {
		return xyResponse.Status{}, err
	}
	result := xyResponse.Status{Database: "postgresql"}
	db := infra.DB.WithContext(ctx)
	for _, count := range []struct {
		model any
		query string
		args  []any
		value *int64
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
