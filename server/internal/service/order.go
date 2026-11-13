package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Order 管理闲鱼订单同步、查询和虚拟发货确认。
type Order struct {
	db      *gorm.DB        // db 持久化订单及订单事件。
	events  *BusinessEvent  // events 记录订单业务结果。
	mtop    mtopCaller      // mtop 调用闲鱼订单与发货接口。
	planner deliveryPlanner // planner 为新订单规划自动发货。
	sender  textSender      // sender 向买家发送发货文本。
}

// NewOrder 创建订单服务。
func NewOrder(db *gorm.DB, events *BusinessEvent) *Order {
	return &Order{db: db, events: events, mtop: &MTopClient{db: db}}
}

// PendingOrderPage 是闲鱼待发货订单接口的业务响应。
type PendingOrderPage struct {
	Module struct {
		Items []struct {
			CommonData struct {
				OrderID        string `json:"orderId"`        // OrderID 是闲鱼订单标识。
				ItemID         string `json:"itemId"`         // ItemID 是闲鱼商品标识。
				OrderStatus    string `json:"orderStatus"`    // OrderStatus 是闲鱼原始订单状态。
				CreateTime     string `json:"createTime"`     // CreateTime 是订单创建时间文本。
				PaySuccessTime string `json:"paySuccessTime"` // PaySuccessTime 是支付成功时间文本。
			} `json:"commonData"` // CommonData 保存订单公共字段。
			BuyerInfo struct {
				UserNick string `json:"userNick"` // UserNick 是买家显示名称。
				UserID   string `json:"userId"`   // UserID 是闲鱼买家标识。
			} `json:"buyerInfoVO"` // BuyerInfo 保存买家信息。
			Item struct {
				Title string `json:"title"` // Title 是成交商品标题。
			} `json:"itemVO"` // Item 保存成交商品信息。
			Price struct {
				TotalPrice string `json:"totalPrice"` // TotalPrice 是订单金额文本。
			} `json:"priceVO"` // Price 保存订单价格信息。
		} `json:"items"` // Items 是当前页待发货订单。
	} `json:"module"` // Module 包含订单列表模块。
}

// caller 返回注入的 MTop 调用器或默认实现。
func (s *Order) caller() mtopCaller {
	if s.mtop != nil {
		return s.mtop
	}
	return &MTopClient{db: s.db}
}

// SyncPending 同步账号待发货订单并为每条新订单规划自动发货。
func (s *Order) SyncPending(ctx context.Context, ownerID, accountID uint) (int, error) {
	if _, err := ownedAccount(ctx, s.db, ownerID, accountID); err != nil {
		return 0, err
	}
	data, err := s.caller().Call(ctx, ownerID, accountID, "mtop.taobao.idle.trade.merchant.sold.get", map[string]any{
		"pageNumber": 1, "rowsPerPage": 100, "orderIds": "", "queryCode": "NOT_SHIP", "orderSearchParam": "{}",
	}, true)
	if err != nil {
		return 0, err
	}
	var page PendingOrderPage
	if err := json.Unmarshal(data, &page); err != nil {
		return 0, errors.New("闲鱼订单响应格式无法识别")
	}
	db := s.db.WithContext(ctx)
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
	s.events.Record(ctx, model.BusinessEvent{OwnerID: ownerID, AccountID: accountID, Type: "order_sync", Status: "success", Summary: fmt.Sprintf("同步 %d 条待发货订单", count)})
	return count, nil
}

// SendDelivery 向买家发送发货内容。
func (s *Order) SendDelivery(ctx context.Context, ownerID uint, item model.Order, text string) error {
	if s.sender == nil {
		return errors.New("闲鱼消息发送服务不可用")
	}
	if err := s.sender.SendText(ctx, ownerID, item.AccountID, item.SessionID, item.BuyerUserID, text); err != nil {
		if errors.Is(err, errSendUnknown) {
			return fmt.Errorf("%w: 发送发货消息: %v", errMTopTransport, err)
		}
		return fmt.Errorf("发送发货消息: %w", err)
	}
	return nil
}

// ConfirmDelivery 确认平台发货。
func (s *Order) ConfirmDelivery(ctx context.Context, ownerID uint, item model.Order, text string) error {
	_, err := s.caller().Call(ctx, ownerID, item.AccountID, "mtop.taobao.idle.logistic.consign.dummy", map[string]any{
		"orderId": item.ExternalOrderID, "tradeText": text, "picList": []any{}, "newUnconsign": true,
	}, false)
	if err != nil && strings.Contains(err.Error(), "ORDER_ALREADY_DELIVERY") {
		return nil
	}
	return err
}

// List 分页查询 ownerID 名下订单并补充账号显示名称。
func (s *Order) List(ctx context.Context, ownerID uint, input xyRequest.OrderSearch) ([]xyResponse.Order, int64, error) {
	if ownerID == 0 {
		return nil, 0, errors.New("用户未登录")
	}
	db := s.db.WithContext(ctx).Table("xy_orders AS o").Where("o.owner_id = ? AND o.deleted_at IS NULL", ownerID)
	if input.AccountID != 0 {
		db = db.Where("o.account_id = ?", input.AccountID)
	}
	if input.ProductID != 0 {
		db = db.Where("o.product_id = ?", input.ProductID)
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
	err := orderDetails(db).
		Scopes(input.Paginate()).Order("o.updated_at DESC, o.id DESC").Scan(&list).Error
	return list, total, err
}

// Get 查询 ownerID 名下的单个订单详情。
func (s *Order) Get(ctx context.Context, ownerID, orderID uint) (xyResponse.Order, error) {
	if ownerID == 0 || orderID == 0 {
		return xyResponse.Order{}, errors.New("订单 ID 不能为空")
	}
	var item xyResponse.Order
	db := s.db.WithContext(ctx).Table("xy_orders AS o").Where("o.id = ? AND o.owner_id = ? AND o.deleted_at IS NULL", orderID, ownerID)
	if err := orderDetails(db).Take(&item).Error; err != nil {
		return xyResponse.Order{}, err
	}
	return item, nil
}

func orderDetails(db *gorm.DB) *gorm.DB {
	return db.Select(`o.id, o.created_at, o.updated_at, o.account_id, a.label AS account_label,
		o.external_order_id, o.external_product_id, o.product_title, o.buyer_user_id, o.buyer_user_name,
		o.session_id, o.sku_key, o.status, o.amount_cent`).
		Joins("JOIN xy_accounts AS a ON a.id = o.account_id AND a.deleted_at IS NULL")
}

// UpsertFromMessage 从订单消息幂等更新本地订单并触发发货规划。
func (s *Order) UpsertFromMessage(ctx context.Context, account model.Account, item model.Message) (*model.Order, error) {
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
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
