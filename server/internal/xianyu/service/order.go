package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model"
	xyRequest "github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model/request"
	xyResponse "github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type order struct {
	mtop    mtopCaller
	planner deliveryPlanner
	sender  textSender
}

type pendingOrderPage struct {
	Module struct {
		Items []struct {
			CommonData struct {
				OrderID        string `json:"orderId"`
				ItemID         string `json:"itemId"`
				OrderStatus    string `json:"orderStatus"`
				CreateTime     string `json:"createTime"`
				PaySuccessTime string `json:"paySuccessTime"`
			} `json:"commonData"`
			BuyerInfo struct {
				UserNick string `json:"userNick"`
				UserID   string `json:"userId"`
			} `json:"buyerInfoVO"`
			Item struct {
				Title string `json:"title"`
			} `json:"itemVO"`
			Price struct {
				TotalPrice string `json:"totalPrice"`
			} `json:"priceVO"`
		} `json:"items"`
	} `json:"module"`
}

func (s *order) caller() mtopCaller {
	if s.mtop != nil {
		return s.mtop
	}
	return new(mtopClient)
}

func (s *order) SyncPending(ctx context.Context, ownerID, accountID uint) (int, error) {
	if _, err := ownedAccount(ctx, ownerID, accountID); err != nil {
		return 0, err
	}
	data, err := s.caller().Call(ctx, ownerID, accountID, "mtop.taobao.idle.trade.merchant.sold.get", map[string]any{
		"pageNumber": 1, "rowsPerPage": 100, "orderIds": "", "queryCode": "NOT_SHIP", "orderSearchParam": "{}",
	}, true)
	if err != nil {
		return 0, err
	}
	var page pendingOrderPage
	if err := json.Unmarshal(data, &page); err != nil {
		return 0, errors.New("闲鱼订单响应格式无法识别")
	}
	db := global.GVA_DB.WithContext(ctx)
	count := 0
	for _, source := range page.Module.Items {
		if strings.TrimSpace(source.CommonData.OrderID) == "" {
			continue
		}
		var product model.Product
		_ = db.Where("owner_id = ? AND account_id = ? AND external_product_id = ?", ownerID, accountID, source.CommonData.ItemID).First(&product).Error
		amount, _ := strconv.ParseFloat(strings.TrimSpace(source.Price.TotalPrice), 64)
		record := model.Order{
			OwnerID: ownerID, AccountID: accountID, ExternalOrderID: source.CommonData.OrderID,
			ProductID: product.ID, ExternalProductID: source.CommonData.ItemID, ProductTitle: source.Item.Title,
			BuyerUserID: source.BuyerInfo.UserID, BuyerUserName: source.BuyerInfo.UserNick,
			Status: model.OrderStatusPendingDelivery, AmountCent: int64(math.Round(amount * 100)),
		}
		if record.ProductTitle == "" {
			record.ProductTitle = product.Title
		}
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "account_id"}, {Name: "external_order_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"product_id", "external_product_id", "product_title", "buyer_user_id", "buyer_user_name", "status", "amount_cent", "updated_at"}),
		}).Create(&record).Error; err != nil {
			return count, err
		}
		if err := db.Where("account_id = ? AND external_order_id = ?", accountID, source.CommonData.OrderID).First(&record).Error; err != nil {
			return count, err
		}
		if s.planner != nil {
			if err := s.planner.Plan(ctx, record); err != nil {
				return count, err
			}
		}
		count++
	}
	recordBusinessEvent(ctx, model.BusinessEvent{OwnerID: ownerID, AccountID: accountID, Type: "order_sync", Status: "success", Summary: fmt.Sprintf("同步 %d 条待发货订单", count)})
	return count, nil
}

func (s *order) Consign(ctx context.Context, ownerID uint, item model.Order, text string) error {
	if s.sender == nil {
		return errors.New("闲鱼消息发送服务不可用")
	}
	if err := s.sender.SendText(ctx, ownerID, item.AccountID, item.SessionID, item.BuyerUserID, text); err != nil {
		if strings.Contains(err.Error(), "结果未知") {
			return fmt.Errorf("%w: 发送发货消息: %v", errMTopTransport, err)
		}
		return fmt.Errorf("发送发货消息: %w", err)
	}
	_, err := s.caller().Call(ctx, ownerID, item.AccountID, "mtop.taobao.idle.logistic.consign.dummy", map[string]any{
		"orderId": item.ExternalOrderID, "tradeText": text, "picList": []any{}, "newUnconsign": true,
	}, false)
	if err != nil && strings.Contains(err.Error(), "ORDER_ALREADY_DELIVERY") {
		return nil
	}
	return err
}

func (s *order) List(ctx context.Context, ownerID uint, input xyRequest.OrderSearch) ([]xyResponse.Order, int64, error) {
	if ownerID == 0 {
		return nil, 0, errors.New("用户未登录")
	}
	db := global.GVA_DB.WithContext(ctx).Table("xy_orders AS o").Where("o.owner_id = ? AND o.deleted_at IS NULL", ownerID)
	if input.AccountID != 0 {
		db = db.Where("o.account_id = ?", input.AccountID)
	}
	if status := strings.TrimSpace(input.Status); status != "" {
		db = db.Where("o.status = ?", status)
	}
	if keyword := strings.TrimSpace(input.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("o.external_order_id LIKE ? OR o.product_title LIKE ? OR o.buyer_user_name LIKE ?", like, like, like)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	list := make([]xyResponse.Order, 0)
	err := db.Select(`o.id, o.created_at, o.updated_at, o.account_id, a.label AS account_label,
		o.external_order_id, o.external_product_id, o.product_title, o.buyer_user_id, o.buyer_user_name,
		o.session_id, o.sku_key, o.status, o.amount_cent`).
		Joins("JOIN xy_accounts AS a ON a.id = o.account_id AND a.deleted_at IS NULL").
		Scopes(input.Paginate()).Order("o.updated_at DESC, o.id DESC").Scan(&list).Error
	return list, total, err
}

func (s *order) UpsertFromMessage(ctx context.Context, account model.Account, item model.Message) (*model.Order, error) {
	if item.ExternalOrderID == "" {
		return nil, nil
	}
	status := model.OrderStatusUnknown
	eventType := "message"
	if item.ContentType == 26 {
		status = model.OrderStatusPendingDelivery
		eventType = model.OrderStatusPendingDelivery
	}
	var saved model.Order
	err := global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var product model.Product
		if err := tx.Where("owner_id = ? AND account_id = ? AND external_product_id = ?", account.OwnerID, account.ID, item.ExternalProductID).
			First(&product).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		orderRecord := model.Order{
			OwnerID: account.OwnerID, AccountID: account.ID, ExternalOrderID: item.ExternalOrderID,
			ProductID: product.ID, ExternalProductID: item.ExternalProductID, ProductTitle: product.Title,
			BuyerUserID: item.SenderUserID, BuyerUserName: item.SenderUserName, SessionID: item.SessionID, Status: status,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "account_id"}, {Name: "external_order_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"product_id": product.ID, "external_product_id": item.ExternalProductID, "product_title": product.Title,
				"buyer_user_id": item.SenderUserID, "buyer_user_name": item.SenderUserName, "session_id": item.SessionID,
			}),
		}).Create(&orderRecord).Error; err != nil {
			return err
		}
		if err := tx.Where("account_id = ? AND external_order_id = ?", account.ID, item.ExternalOrderID).First(&orderRecord).Error; err != nil {
			return err
		}
		if status == model.OrderStatusPendingDelivery && orderRecord.Status == model.OrderStatusUnknown {
			if err := tx.Model(&orderRecord).Update("status", status).Error; err != nil {
				return err
			}
			orderRecord.Status = status
		}
		saved = orderRecord
		event := model.OrderEvent{
			OwnerID: account.OwnerID, AccountID: account.ID, OrderID: orderRecord.ID,
			PNMID: item.PNMID, Type: eventType, EventTime: item.MessageTime,
		}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&event).Error
	})
	return &saved, err
}
