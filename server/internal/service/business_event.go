package service

import (
	"context"
	"errors"
	"strings"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/response"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BusinessEvent struct {
	db *gorm.DB
}

func NewBusinessEvent(db *gorm.DB) *BusinessEvent {
	return &BusinessEvent{db: db}
}

func (s *BusinessEvent) Record(ctx context.Context, event model.BusinessEvent) {
	fields := []zap.Field{
		zap.Uint("ownerID", event.OwnerID), zap.Uint("accountID", event.AccountID),
		zap.Uint("productID", event.ProductID), zap.Uint("orderID", event.OrderID),
		zap.String("type", event.Type), zap.String("status", event.Status),
		zap.String("summary", event.Summary),
	}
	if event.Error != "" {
		fields = append(fields, zap.String("error", event.Error))
	}
	if event.Status == "success" || event.Status == "ready" {
		zap.L().Info("闲鱼业务操作完成", fields...)
	} else {
		zap.L().Warn("闲鱼业务操作异常", fields...)
	}
	if event.OwnerID != 0 {
		if err := s.db.WithContext(ctx).Create(&event).Error; err != nil {
			zap.L().Error("保存闲鱼业务日志失败", zap.Error(err), zap.String("type", event.Type), zap.Uint("accountID", event.AccountID))
		}
	}
}

func (s *BusinessEvent) List(ctx context.Context, ownerID uint, input xyRequest.BusinessEventSearch) ([]xyResponse.BusinessEvent, int64, error) {
	if ownerID == 0 {
		return nil, 0, errors.New("用户未登录")
	}
	db := s.db.WithContext(ctx).Table("xy_business_events AS e").Where("e.owner_id = ? AND e.deleted_at IS NULL", ownerID)
	if input.AccountID != 0 {
		db = db.Where("e.account_id = ?", input.AccountID)
	}
	if input.ProductID != 0 {
		db = db.Where("e.product_id = ?", input.ProductID)
	}
	if input.OrderID != 0 {
		db = db.Where("e.order_id = ?", input.OrderID)
	}
	if input.Type != "" {
		db = db.Where("e.type = ?", input.Type)
	}
	if input.Status != "" {
		db = db.Where("e.status = ?", input.Status)
	}
	if keyword := strings.TrimSpace(input.Keyword); keyword != "" {
		db = db.Where("e.summary LIKE ? OR e.error LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	list := make([]xyResponse.BusinessEvent, 0)
	err := db.Select(`e.id, e.created_at, e.account_id, COALESCE(a.label, '') AS account_label,
		e.product_id, COALESCE(p.title, '') AS product_title, e.order_id, COALESCE(o.external_order_id, '') AS external_order_id,
		e.type, e.status, e.summary, e.error`).
		Joins("LEFT JOIN xy_accounts AS a ON a.id = e.account_id").
		Joins("LEFT JOIN xy_products AS p ON p.id = e.product_id").
		Joins("LEFT JOIN xy_orders AS o ON o.id = e.order_id").
		Scopes(input.Paginate()).Order("e.id DESC").Scan(&list).Error
	return list, total, err
}
