package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xianyu/model"
	xianyuRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/xianyu/model/request"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type relist struct {
	mtop     mtopCaller
	products *product
	media    func(context.Context, uint, uint, map[string]any) error
}

func (s *relist) caller() mtopCaller {
	if s.mtop != nil {
		return s.mtop
	}
	return new(mtopClient)
}

func (s *relist) Plan(ctx context.Context, order model.Order) error {
	if order.Status != model.OrderStatusShipped || order.ProductID == 0 {
		return nil
	}
	var setting model.ProductSetting
	if err := global.GVA_DB.WithContext(ctx).Where("product_id = ? AND auto_relist_enabled = ?", order.ProductID, true).First(&setting).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	var item model.Product
	if err := global.GVA_DB.WithContext(ctx).Where("id = ? AND owner_id = ?", order.ProductID, order.OwnerID).First(&item).Error; err != nil {
		return err
	}
	if strings.TrimSpace(item.PublishSnapshot) == "" {
		return errors.New("商品没有可用的发布快照")
	}
	task := model.RelistTask{
		OwnerID: order.OwnerID, AccountID: order.AccountID, OrderID: order.ID, ProductID: order.ProductID,
		Status: model.RelistTaskPending, NextAttemptAt: time.Now().Add(time.Duration(setting.RelistDelaySeconds) * time.Second),
	}
	return global.GVA_DB.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&task).Error
}

func (s *relist) ProcessTask(ctx context.Context, ownerID, taskID uint) error {
	db := global.GVA_DB.WithContext(ctx)
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
		if _, err := s.products.Sync(ctx, ownerID, xianyuRequest.ProductSync{AccountID: task.AccountID}); err != nil {
			return s.failTask(db, task, fmt.Errorf("确认原商品状态: %w", err))
		}
		if err := db.First(&item, task.ProductID).Error; err != nil {
			return s.failTask(db, task, err)
		}
		if item.Status == "on_sale" {
			return db.Model(&task).Updates(map[string]any{"status": model.RelistTaskPending, "next_attempt_at": time.Now().Add(30 * time.Second)}).Error
		}
	}
	var snapshot publishSnapshot
	if err := json.Unmarshal([]byte(item.PublishSnapshot), &snapshot); err != nil || snapshot.API == "" || len(snapshot.Payload) == 0 {
		return s.failTask(db, task, errors.New("商品发布快照已损坏"))
	}
	if err := completePublishCategory(ctx, s.caller(), ownerID, task.AccountID, snapshot.Payload); err != nil {
		return s.failTask(db, task, err)
	}
	snapshot.Payload["uniqueCode"] = strconv.FormatInt(time.Now().UnixMilli(), 10) + strconv.FormatUint(uint64(task.AccountID), 10)
	uploadImages := s.media
	if uploadImages == nil {
		uploadImages = reuploadPublishImages
	}
	if err := uploadImages(ctx, ownerID, task.AccountID, snapshot.Payload); err != nil {
		return s.failTask(db, task, err)
	}
	data, err := s.caller().Call(ctx, ownerID, task.AccountID, snapshot.API, snapshot.Payload, snapshot.Seller)
	if err != nil {
		if errors.Is(err, errMTopTransport) {
			recordBusinessEvent(ctx, model.BusinessEvent{OwnerID: ownerID, AccountID: task.AccountID, ProductID: task.ProductID, OrderID: task.OrderID, Type: "relist", Status: "unknown", Summary: "自动续售结果未知", Error: err.Error()})
			return db.Model(&task).Updates(map[string]any{"status": model.RelistTaskUnknown, "last_error": err.Error()}).Error
		}
		return s.failTask(db, task, err)
	}
	newItemID := findPublishedItemID(data)
	if newItemID == "" {
		err := errors.New("闲鱼已接收发布请求但未返回新商品 ID")
		recordBusinessEvent(ctx, model.BusinessEvent{OwnerID: ownerID, AccountID: task.AccountID, ProductID: task.ProductID, OrderID: task.OrderID, Type: "relist", Status: "unknown", Summary: "自动续售结果未知", Error: err.Error()})
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
		recordBusinessEvent(ctx, model.BusinessEvent{OwnerID: ownerID, AccountID: task.AccountID, ProductID: newProductID, OrderID: task.OrderID, Type: "relist", Status: "success", Summary: "商品自动续售成功"})
	}
	return txErr
}

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

func firstNonEmpty(values ...any) string {
	for _, value := range values {
		if text := strings.TrimSpace(fmt.Sprint(value)); text != "" && text != "<nil>" {
			return text
		}
	}
	return ""
}

func (s *relist) processDue(ctx context.Context, accountID uint) {
	var tasks []model.RelistTask
	if err := global.GVA_DB.WithContext(ctx).Where("account_id = ? AND status IN ? AND next_attempt_at <= ?", accountID,
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

func (s *relist) failTask(db *gorm.DB, task model.RelistTask, err error) error {
	delay := time.Duration(1<<min(task.Attempts, 6)) * time.Second
	updateErr := db.Model(&task).Updates(map[string]any{"status": model.RelistTaskFailed, "last_error": err.Error(), "next_attempt_at": time.Now().Add(delay)}).Error
	recordBusinessEvent(db.Statement.Context, model.BusinessEvent{OwnerID: task.OwnerID, AccountID: task.AccountID, ProductID: task.ProductID, OrderID: task.OrderID, Type: "relist", Status: "failed", Summary: "自动续售失败", Error: err.Error()})
	return errors.Join(err, updateErr)
}

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
