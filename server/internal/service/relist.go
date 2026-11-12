package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Relist 管理售出商品的自动重发计划与执行。
type Relist struct {
	db       *gorm.DB                                                // db 持久化续售任务和商品快照。
	events   *BusinessEvent                                          // events 记录续售业务结果。
	mtop     mtopCaller                                              // mtop 查询分类并发布新商品。
	products *Product                                                // products 同步商品并刷新发布快照。
	media    func(context.Context, uint, uint, map[string]any) error // media 重新上传发布图片。
}

// NewRelist 创建自动续售服务。
func NewRelist(db *gorm.DB, events *BusinessEvent, products *Product) *Relist {
	return &Relist{db: db, events: events, mtop: &MTopClient{db: db}, products: products}
}

// caller 返回注入的 MTop 调用器或默认实现。
func (s *Relist) caller() mtopCaller {
	if s.mtop != nil {
		return s.mtop
	}
	return &MTopClient{db: s.db}
}

// Plan 为已发货且启用自动续售的订单幂等创建延迟任务。
func (s *Relist) Plan(ctx context.Context, order model.Order) error {
	if order.Status != model.OrderStatusShipped || order.ProductID == 0 {
		return nil
	}
	var setting model.ProductSetting
	if err := s.db.WithContext(ctx).Where("product_id = ? AND auto_relist_enabled = ?", order.ProductID, true).First(&setting).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	var item model.Product
	if err := s.db.WithContext(ctx).Where("id = ? AND owner_id = ?", order.ProductID, order.OwnerID).First(&item).Error; err != nil {
		return err
	}
	if s.products != nil {
		snapshot, err := s.products.capturePublishSnapshot(ctx, order.OwnerID, item, xyRequest.UpdateProductSetting{
			ProductID: item.ID, AutoRelistEnabled: true,
			RelistDescription: setting.RelistDescription, RelistImageURLs: setting.RelistImageURLs, RelistAddress: setting.RelistAddress,
		})
		if err != nil {
			return fmt.Errorf("刷新续售发布快照: %w", err)
		}
		item.PublishSnapshot = snapshot
		if err := s.db.WithContext(ctx).Model(&item).Update("publish_snapshot", snapshot).Error; err != nil {
			return err
		}
	}
	if strings.TrimSpace(item.PublishSnapshot) == "" {
		return errors.New("商品没有可用的发布快照")
	}
	task := model.RelistTask{
		OwnerID: order.OwnerID, AccountID: order.AccountID, OrderID: order.ID, ProductID: order.ProductID,
		Status: model.RelistTaskPending, NextAttemptAt: time.Now().Add(time.Duration(setting.RelistDelaySeconds) * time.Second),
	}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&task).Error
}

// ListTasks 分页查询 ownerID 名下的自动续售任务。
func (s *Relist) ListTasks(ctx context.Context, ownerID uint, input xyRequest.RelistTaskSearch) ([]model.RelistTask, int64, error) {
	if ownerID == 0 {
		return nil, 0, errors.New("用户未登录")
	}
	db := s.db.WithContext(ctx).Model(&model.RelistTask{}).Where("owner_id = ?", ownerID)
	if input.AccountID != 0 {
		db = db.Where("account_id = ?", input.AccountID)
	}
	if input.ProductID != 0 {
		db = db.Where("product_id = ?", input.ProductID)
	}
	if status := strings.TrimSpace(input.Status); status != "" {
		db = db.Where("status = ?", status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	tasks := make([]model.RelistTask, 0)
	err := db.Scopes(input.Paginate()).Order("id DESC").Find(&tasks).Error
	return tasks, total, err
}

// Retry 仅重试 ownerID 名下明确失败的续售任务。
func (s *Relist) Retry(ctx context.Context, ownerID, taskID uint) error {
	if ownerID == 0 || taskID == 0 {
		return errors.New("续售任务 ID 不能为空")
	}
	result := s.db.WithContext(ctx).Model(&model.RelistTask{}).
		Where("id = ? AND owner_id = ? AND status = ?", taskID, ownerID, model.RelistTaskFailed).
		Updates(map[string]any{"status": model.RelistTaskPending, "next_attempt_at": time.Now(), "last_error": ""})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errors.New("只有明确失败的续售任务可以重试；结果未知时请先核实闲鱼状态")
	}
	return s.ProcessTask(ctx, ownerID, taskID)
}

// ProcessTask 原子认领续售任务，确认原商品状态后发布新商品并迁移设置。
func (s *Relist) ProcessTask(ctx context.Context, ownerID, taskID uint) error {
	db := s.db.WithContext(ctx)
	claim := db.Model(&model.RelistTask{}).
		Where("id = ? AND owner_id = ? AND status IN ? AND next_attempt_at <= ?", taskID, ownerID,
			[]string{model.RelistTaskPending, model.RelistTaskFailed}, time.Now()).
		Updates(map[string]any{"status": model.RelistTaskPublishing, "attempts": gorm.Expr("attempts + 1"), "last_error": ""})
	if claim.Error != nil {
		return claim.Error
	}
	if claim.RowsAffected != 1 {
		return nil
	}
	var task model.RelistTask
	if err := db.First(&task, taskID).Error; err != nil {
		return err
	}
	var item model.Product
	if err := db.Where("id = ? AND owner_id = ?", task.ProductID, ownerID).First(&item).Error; err != nil {
		return s.failTask(db, task, err)
	}
	if s.products != nil {
		if _, err := s.products.Sync(ctx, ownerID, xyRequest.ProductSync{AccountID: task.AccountID}); err != nil {
			return s.failTask(db, task, fmt.Errorf("确认原商品状态: %w", err))
		}
		if err := db.First(&item, task.ProductID).Error; err != nil {
			return s.failTask(db, task, err)
		}
		if item.Status == "on_sale" {
			return db.Model(&task).Updates(map[string]any{"status": model.RelistTaskPending, "next_attempt_at": time.Now().Add(30 * time.Second)}).Error
		}
	}
	var snapshot PublishSnapshot
	if err := json.Unmarshal([]byte(item.PublishSnapshot), &snapshot); err != nil || snapshot.API == "" || len(snapshot.Payload) == 0 {
		return s.failTask(db, task, errors.New("商品发布快照已损坏"))
	}
	if err := completePublishCategory(ctx, s.caller(), ownerID, task.AccountID, snapshot.Payload); err != nil {
		return s.failTask(db, task, err)
	}
	snapshot.Payload["uniqueCode"] = strconv.FormatInt(time.Now().UnixMilli(), 10) + strconv.FormatUint(uint64(task.AccountID), 10)
	uploadImages := s.media
	if uploadImages == nil {
		uploadImages = func(ctx context.Context, ownerID, accountID uint, payload map[string]any) error {
			return reuploadPublishImages(ctx, s.db, ownerID, accountID, payload)
		}
	}
	if err := uploadImages(ctx, ownerID, task.AccountID, snapshot.Payload); err != nil {
		return s.failTask(db, task, err)
	}
	data, err := s.caller().Call(ctx, ownerID, task.AccountID, snapshot.API, snapshot.Payload, snapshot.Seller)
	if err != nil {
		if errors.Is(err, errMTopTransport) {
			s.events.Record(ctx, model.BusinessEvent{OwnerID: ownerID, AccountID: task.AccountID, ProductID: task.ProductID, OrderID: task.OrderID, Type: "relist", Status: "unknown", Summary: "自动续售结果未知", Error: err.Error()})
			return db.Model(&task).Updates(map[string]any{"status": model.RelistTaskUnknown, "last_error": err.Error()}).Error
		}
		return s.failTask(db, task, err)
	}
	newItemID := findPublishedItemID(data)
	if newItemID == "" {
		err := errors.New("闲鱼已接收发布请求但未返回新商品 ID")
		s.events.Record(ctx, model.BusinessEvent{OwnerID: ownerID, AccountID: task.AccountID, ProductID: task.ProductID, OrderID: task.OrderID, Type: "relist", Status: "unknown", Summary: "自动续售结果未知", Error: err.Error()})
		return db.Model(&task).Updates(map[string]any{"status": model.RelistTaskUnknown, "last_error": err.Error()}).Error
	}
	var newProductID uint
	txErr := db.Transaction(func(tx *gorm.DB) error {
		newItem := model.Product{OwnerID: item.OwnerID, AccountID: item.AccountID, ExternalProductID: newItemID, Title: item.Title, PriceCent: item.PriceCent, Status: "on_sale", ImageURL: item.ImageURL, PublishSnapshot: item.PublishSnapshot, SourceJSON: item.SourceJSON}
		if err := tx.Create(&newItem).Error; err != nil {
			return err
		}
		newProductID = newItem.ID
		for _, target := range []any{&model.ProductSetting{}, &model.DeliveryRule{}, &model.ReplyRule{}} {
			if err := tx.Model(target).Where("product_id = ?", item.ID).Update("product_id", newItem.ID).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&item).Update("status", "sold").Error; err != nil {
			return err
		}
		if err := tx.Model(&task).Updates(map[string]any{"status": model.RelistTaskSucceeded, "new_external_product_id": newItemID, "last_error": ""}).Error; err != nil {
			return err
		}
		return nil
	})
	if txErr == nil {
		s.events.Record(ctx, model.BusinessEvent{OwnerID: ownerID, AccountID: task.AccountID, ProductID: newProductID, OrderID: task.OrderID, Type: "relist", Status: "success", Summary: "商品自动续售成功"})
	}
	return txErr
}

// completePublishCategory 在快照缺少发布分类路径时向闲鱼补查并写回 payload。
func completePublishCategory(ctx context.Context, caller mtopCaller, ownerID, accountID uint, payload map[string]any) error {
	category, ok := payload["itemCatDTO"].(map[string]any)
	if !ok {
		return nil
	}
	catID := strings.TrimSpace(fmt.Sprint(category["catId"]))
	if catID == "" || catID == "<nil>" {
		return errors.New("续售快照缺少闲鱼分类 ID")
	}
	if strings.TrimSpace(fmt.Sprint(category["channelCatId"])) != "" && strings.TrimSpace(fmt.Sprint(category["tbCatId"])) != "" {
		return nil
	}
	text, _ := payload["itemTextDTO"].(map[string]any)
	title := strings.TrimSpace(fmt.Sprint(text["title"]))
	description := strings.TrimSpace(fmt.Sprint(text["desc"]))
	data, err := caller.Call(ctx, ownerID, accountID, "mtop.taobao.idle.kgraph.pc.property.recommend/2.0", map[string]any{
		"title": title, "description": description, "lockCpv": false, "multiSKU": false,
		"publishScene": "pcBackendPublish", "scene": "shopPcPublish",
		"uniqueCode": strconv.FormatInt(time.Now().UnixMilli(), 10),
	}, true)
	if err != nil {
		return fmt.Errorf("查询闲鱼商品分类: %w", err)
	}
	fields, err := findRecommendedCategory(data, catID)
	if err != nil {
		return err
	}
	for key, value := range fields {
		if value != "" {
			category[key] = value
		}
	}
	if strings.TrimSpace(fmt.Sprint(category["channelCatId"])) == "" || strings.TrimSpace(fmt.Sprint(category["tbCatId"])) == "" {
		return errors.New("闲鱼分类推荐结果未包含原商品所需分类路径")
	}
	return nil
}

// findRecommendedCategory 递归查找与 targetCatID 匹配且路径完整的分类。
func findRecommendedCategory(raw json.RawMessage, targetCatID string) (map[string]string, error) {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, errors.New("闲鱼分类推荐结果无法解析")
	}
	var find func(any) map[string]string
	find = func(value any) map[string]string {
		switch current := value.(type) {
		case map[string]any:
			catID := firstNonEmpty(current["catId"], current["categoryId"])
			channelCatID := firstNonEmpty(current["channelCatId"], current["channelCateId"], current["channelCategoryId"])
			tbCatID := firstNonEmpty(current["tbCatId"], current["taobaoCategoryId"])
			if catID == targetCatID && channelCatID != "" && tbCatID != "" {
				return map[string]string{
					"catId": catID, "catName": firstNonEmpty(current["catName"], current["categoryName"]),
					"channelCatId": channelCatID, "tbCatId": tbCatID,
				}
			}
			for _, child := range current {
				if result := find(child); result != nil {
					return result
				}
			}
		case []any:
			for _, child := range current {
				if result := find(child); result != nil {
					return result
				}
			}
		case string:
			var nested any
			if json.Unmarshal([]byte(current), &nested) == nil {
				return find(nested)
			}
		}
		return nil
	}
	if fields := find(value); fields != nil {
		return fields, nil
	}
	return nil, errors.New("闲鱼分类推荐未找到与原商品匹配的分类路径")
}

// firstNonEmpty 返回第一个非空且非 nil 文本值。
func firstNonEmpty(values ...any) string {
	for _, value := range values {
		if text := strings.TrimSpace(fmt.Sprint(value)); text != "" && text != "<nil>" {
			return text
		}
	}
	return ""
}

// processDue 顺序处理账号最多十个到期续售任务。
func (s *Relist) processDue(ctx context.Context, accountID uint) {
	var tasks []model.RelistTask
	if err := s.db.WithContext(ctx).Where("account_id = ? AND status IN ? AND next_attempt_at <= ?", accountID,
		[]string{model.RelistTaskPending, model.RelistTaskFailed}, time.Now()).Order("id").Limit(10).Find(&tasks).Error; err != nil {
		return
	}
	for _, task := range tasks {
		if ctx.Err() != nil {
			return
		}
		_ = s.ProcessTask(ctx, task.OwnerID, task.ID)
	}
}

// failTask 记录失败并按尝试次数设置指数退避时间。
func (s *Relist) failTask(db *gorm.DB, task model.RelistTask, err error) error {
	delay := time.Duration(1<<min(task.Attempts, 6)) * time.Second
	updateErr := db.Model(&task).Updates(map[string]any{"status": model.RelistTaskFailed, "last_error": err.Error(), "next_attempt_at": time.Now().Add(delay)}).Error
	s.events.Record(db.Statement.Context, model.BusinessEvent{OwnerID: task.OwnerID, AccountID: task.AccountID, ProductID: task.ProductID, OrderID: task.OrderID, Type: "relist", Status: "failed", Summary: "自动续售失败", Error: err.Error()})
	return errors.Join(err, updateErr)
}

// findPublishedItemID 递归提取发布响应中的新商品标识。
func findPublishedItemID(data json.RawMessage) string {
	var value any
	if json.Unmarshal(data, &value) != nil {
		return ""
	}
	var find func(any) string
	find = func(value any) string {
		switch current := value.(type) {
		case map[string]any:
			for _, key := range []string{"itemId", "item_id"} {
				if text := strings.TrimSpace(fmt.Sprint(current[key])); text != "" && text != "<nil>" {
					return text
				}
			}
			for _, child := range current {
				if result := find(child); result != "" {
					return result
				}
			}
		case []any:
			for _, child := range current {
				if result := find(child); result != "" {
					return result
				}
			}
		}
		return ""
	}
	return find(value)
}
