package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Delivery 管理卡密库存、发货任务和账号级后台执行器。
type Delivery struct {
	root      context.Context // root 控制账号后台执行器生命周期。
	db        *gorm.DB        // db 持久化规则、卡密和发货任务。
	events    *BusinessEvent  // events 记录发货业务结果。
	consigner interface {
		// Consign 向买家发送内容并确认闲鱼订单发货。
		Consign(context.Context, uint, model.Order, string) error
	} // consigner 执行不可重复的远端发货动作。
	rules   *DeliveryRule               // rules 为订单选择发货模板。
	relist  *Relist                     // relist 在发货成功后规划自动续售。
	mu      sync.Mutex                  // mu 保护 workers。
	workers map[uint]context.CancelFunc // workers 保存账号后台执行器取消函数。
}

// NewDelivery 创建自动发货服务并初始化账号执行器集合。
func NewDelivery(ctx context.Context, db *gorm.DB, events *BusinessEvent, orders *Order, rules *DeliveryRule, relist *Relist) *Delivery {
	return &Delivery{root: contextOrBackground(ctx), db: db, events: events, consigner: orders, rules: rules, relist: relist, workers: make(map[uint]context.CancelFunc)}
}

// ImportResources 校验规则归属，去重并加密导入最多一千条卡密。
func (s *Delivery) ImportResources(ctx context.Context, ownerID uint, input xyRequest.ImportDeliveryResources) (int, error) {
	if ownerID == 0 || input.RuleID == 0 || len(input.Values) == 0 || len(input.Values) > 1000 {
		return 0, errors.New("规则和 1 到 1000 条卡密不能为空")
	}
	var rule model.DeliveryRule
	if err := s.db.WithContext(ctx).Where("id = ? AND owner_id = ? AND mode = ?", input.RuleID, ownerID, model.DeliveryModeCard).First(&rule).Error; err != nil {
		return 0, err
	}
	resources := make([]model.DeliveryResource, 0, len(input.Values))
	seen := make(map[string]struct{}, len(input.Values))
	for _, raw := range input.Values {
		value := strings.TrimSpace(raw)
		if value == "" || len([]rune(value)) > 4000 {
			return 0, errors.New("卡密不能为空且不能超过 4000 个字符")
		}
		hashBytes := sha256.Sum256([]byte(value))
		hash := hex.EncodeToString(hashBytes[:])
		if _, exists := seen[hash]; exists {
			continue
		}
		seen[hash] = struct{}{}
		ciphertext, err := encryptCredential(s.db, value, deliveryResourceAAD(rule.ID, hash))
		if err != nil {
			return 0, err
		}
		resources = append(resources, model.DeliveryResource{
			OwnerID: ownerID, RuleID: rule.ID, ValueHash: hash, ValueCiphertext: ciphertext, Status: model.DeliveryResourceAvailable,
		})
	}
	result := s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&resources)
	return int(result.RowsAffected), result.Error
}

// Inventory 按状态汇总 ownerID 名下指定规则的卡密数量。
func (s *Delivery) Inventory(ctx context.Context, ownerID, ruleID uint) (xyResponse.DeliveryInventory, error) {
	if ownerID == 0 || ruleID == 0 {
		return xyResponse.DeliveryInventory{}, errors.New("规则 ID 不能为空")
	}
	var owned int64
	if err := s.db.WithContext(ctx).Model(&model.DeliveryRule{}).Where("id = ? AND owner_id = ?", ruleID, ownerID).Count(&owned).Error; err != nil || owned != 1 {
		if err != nil {
			return xyResponse.DeliveryInventory{}, err
		}
		return xyResponse.DeliveryInventory{}, gorm.ErrRecordNotFound
	}
	result := xyResponse.DeliveryInventory{RuleID: ruleID}
	rows, err := s.db.WithContext(ctx).Model(&model.DeliveryResource{}).
		Select("status, COUNT(*) AS count").Where("owner_id = ? AND rule_id = ?", ownerID, ruleID).Group("status").Rows()
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return result, err
		}
		switch status {
		case model.DeliveryResourceAvailable:
			result.Available = count
		case model.DeliveryResourceReserved:
			result.Reserved = count
		case model.DeliveryResourceConsumed:
			result.Consumed = count
		}
	}
	return result, rows.Err()
}

// ListTasks 分页查询 ownerID 名下发货任务并补充订单、商品和账号信息。
func (s *Delivery) ListTasks(ctx context.Context, ownerID uint, input xyRequest.DeliveryTaskSearch) ([]xyResponse.DeliveryTask, int64, error) {
	if ownerID == 0 {
		return nil, 0, errors.New("用户未登录")
	}
	db := s.db.WithContext(ctx).Table("xy_delivery_tasks AS t").
		Joins("JOIN xy_orders AS o ON o.id = t.order_id AND o.deleted_at IS NULL").
		Where("t.owner_id = ? AND t.deleted_at IS NULL", ownerID)
	if input.AccountID != 0 {
		db = db.Where("t.account_id = ?", input.AccountID)
	}
	if input.ProductID != 0 {
		db = db.Where("o.product_id = ?", input.ProductID)
	}
	if status := strings.TrimSpace(input.Status); status != "" {
		db = db.Where("t.status = ?", status)
	}
	if keyword := strings.TrimSpace(input.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("o.external_order_id LIKE ? OR o.product_title LIKE ? OR o.buyer_user_name LIKE ?", like, like, like)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	list := make([]xyResponse.DeliveryTask, 0)
	err := db.Select(`t.id, t.created_at, t.updated_at, t.account_id, a.label AS account_label,
		o.external_order_id, o.product_title, o.buyer_user_name, r.mode, t.status, t.attempts,
		t.next_attempt_at, t.last_error, t.sent_at, t.confirmed_at`).
		Joins("JOIN xy_accounts AS a ON a.id = t.account_id AND a.deleted_at IS NULL").
		Joins("JOIN xy_delivery_rules AS r ON r.id = t.rule_id AND r.deleted_at IS NULL").
		Scopes(input.Paginate()).Order("t.id DESC").Scan(&list).Error
	return list, total, err
}

// Plan 为符合商品设置的待发货订单幂等创建任务并预留卡密。
func (s *Delivery) Plan(ctx context.Context, order model.Order) error {
	if order.Status != model.OrderStatusPendingDelivery || order.ProductID == 0 || order.ExternalOrderID == "" {
		return nil
	}
	var setting model.ProductSetting
	if err := s.db.WithContext(ctx).Where("product_id = ? AND auto_delivery_enabled = ?", order.ProductID, true).First(&setting).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	rule, err := s.rules.Resolve(ctx, order.OwnerID, order.ProductID, order.SKUKey)
	if err != nil || rule == nil {
		return err
	}
	now := time.Now()
	task := model.DeliveryTask{
		OwnerID: order.OwnerID, AccountID: order.AccountID, OrderID: order.ID, RuleID: rule.ID,
		Status: model.DeliveryTaskPending, NextAttemptAt: now,
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&task).Error; err != nil {
			return err
		}
		if task.ID == 0 {
			return nil
		}
		payload := rule.Template
		if rule.Mode == model.DeliveryModeCard {
			var resource model.DeliveryResource
			if err := tx.Where("owner_id = ? AND rule_id = ? AND status = ?", order.OwnerID, rule.ID, model.DeliveryResourceAvailable).
				Order("id").First(&resource).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					task.Status = model.DeliveryTaskFailed
					return tx.Model(&task).Updates(map[string]any{"status": model.DeliveryTaskFailed, "last_error": "卡密库存不足"}).Error
				}
				return err
			}
			reserve := tx.Model(&model.DeliveryResource{}).
				Where("id = ? AND status = ?", resource.ID, model.DeliveryResourceAvailable).
				Updates(map[string]any{"status": model.DeliveryResourceReserved, "reserved_task_id": task.ID})
			if reserve.Error != nil {
				return reserve.Error
			}
			if reserve.RowsAffected != 1 {
				return errors.New("卡密被其他任务占用")
			}
			value, err := decryptCredential(s.db, resource.ValueCiphertext, deliveryResourceAAD(rule.ID, resource.ValueHash))
			if err != nil {
				return err
			}
			payload = strings.ReplaceAll(rule.Template, "{card}", value)
			task.ResourceID = resource.ID
		}
		ciphertext, err := encryptCredential(s.db, payload, deliveryTaskAAD(order.AccountID, order.ExternalOrderID))
		if err != nil {
			return err
		}
		task.PayloadCiphertext = ciphertext
		return tx.Model(&task).Updates(map[string]any{"resource_id": task.ResourceID, "payload_ciphertext": ciphertext}).Error
	})
	if err == nil && task.ID != 0 && task.Status == model.DeliveryTaskPending {
		go func() { _ = s.ProcessTask(contextOrBackground(s.root), order.OwnerID, task.ID) }()
	}
	return err
}

// ProcessTask 原子认领任务，解密负载并执行一次可审计的远端发货。
func (s *Delivery) ProcessTask(ctx context.Context, ownerID, taskID uint) error {
	if ownerID == 0 || taskID == 0 {
		return errors.New("发货任务 ID 不能为空")
	}
	if s.consigner == nil {
		return errors.New("闲鱼发货服务不可用")
	}
	db := s.db.WithContext(ctx)
	claim := db.Model(&model.DeliveryTask{}).
		Where("id = ? AND owner_id = ? AND status IN ? AND next_attempt_at <= ?", taskID, ownerID,
			[]string{model.DeliveryTaskPending, model.DeliveryTaskFailed}, time.Now()).
		Updates(map[string]any{"status": model.DeliveryTaskSending, "attempts": gorm.Expr("attempts + 1"), "last_error": ""})
	if claim.Error != nil {
		return claim.Error
	}
	if claim.RowsAffected != 1 {
		return errors.New("发货任务当前不可执行")
	}
	var task model.DeliveryTask
	if err := db.First(&task, taskID).Error; err != nil {
		return err
	}
	var order model.Order
	if err := db.Where("id = ? AND owner_id = ?", task.OrderID, ownerID).First(&order).Error; err != nil {
		return s.failTask(db, task, err)
	}
	payload, err := decryptCredential(s.db, task.PayloadCiphertext, deliveryTaskAAD(order.AccountID, order.ExternalOrderID))
	if err != nil {
		return s.failTask(db, task, err)
	}
	if err := s.consigner.Consign(ctx, ownerID, order, payload); err != nil {
		if errors.Is(err, errMTopTransport) {
			s.events.Record(ctx, model.BusinessEvent{OwnerID: ownerID, AccountID: task.AccountID, OrderID: task.OrderID, Type: "delivery", Status: "unknown", Summary: "自动发货结果未知", Error: err.Error()})
			return db.Model(&task).Updates(map[string]any{"status": model.DeliveryTaskUnknown, "last_error": err.Error()}).Error
		}
		return s.failTask(db, task, err)
	}
	now := time.Now()
	txErr := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&task).Updates(map[string]any{"status": model.DeliveryTaskConfirmed, "sent_at": now, "confirmed_at": now, "last_error": ""}).Error; err != nil {
			return err
		}
		if task.ResourceID != 0 {
			if err := tx.Model(&model.DeliveryResource{}).Where("id = ? AND reserved_task_id = ?", task.ResourceID, task.ID).
				Updates(map[string]any{"status": model.DeliveryResourceConsumed, "consumed_at": now}).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&order).Update("status", model.OrderStatusShipped).Error; err != nil {
			return err
		}
		return nil
	})
	if txErr == nil {
		s.events.Record(ctx, model.BusinessEvent{OwnerID: ownerID, AccountID: task.AccountID, ProductID: order.ProductID, OrderID: order.ID, Type: "delivery", Status: "success", Summary: "闲鱼虚拟发货成功"})
		if s.relist != nil {
			order.Status = model.OrderStatusShipped
			if err := s.relist.Plan(ctx, order); err != nil {
				s.events.Record(ctx, model.BusinessEvent{OwnerID: ownerID, AccountID: task.AccountID, ProductID: order.ProductID, OrderID: order.ID, Type: "relist", Status: "failed", Summary: "创建自动续售任务失败", Error: err.Error()})
			}
		}
	}
	return txErr
}

// Retry 仅将明确失败的任务重置为待执行并立即处理。
func (s *Delivery) Retry(ctx context.Context, ownerID, taskID uint) error {
	result := s.db.WithContext(ctx).Model(&model.DeliveryTask{}).
		Where("id = ? AND owner_id = ? AND status = ?", taskID, ownerID, model.DeliveryTaskFailed).
		Updates(map[string]any{"status": model.DeliveryTaskPending, "next_attempt_at": time.Now(), "last_error": ""})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errors.New("只有明确失败的任务可以重试")
	}
	return s.ProcessTask(ctx, ownerID, taskID)
}

// StartAccount 为账号启动唯一的到期任务轮询器。
func (s *Delivery) StartAccount(accountID uint) {
	s.mu.Lock()
	if s.workers == nil {
		s.workers = make(map[uint]context.CancelFunc)
	}
	if s.workers[accountID] != nil {
		s.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(contextOrBackground(s.root))
	s.workers[accountID] = cancel
	s.mu.Unlock()
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			s.processDue(ctx, accountID)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

// StopAccount 停止并移除指定账号的任务轮询器。
func (s *Delivery) StopAccount(accountID uint) {
	s.mu.Lock()
	cancel := s.workers[accountID]
	delete(s.workers, accountID)
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// processDue 顺序处理账号最多十个到期发货任务及续售任务。
func (s *Delivery) processDue(ctx context.Context, accountID uint) {
	var tasks []model.DeliveryTask
	if err := s.db.WithContext(ctx).Where("account_id = ? AND status IN ? AND next_attempt_at <= ?", accountID,
		[]string{model.DeliveryTaskPending, model.DeliveryTaskFailed}, time.Now()).Order("id").Limit(10).Find(&tasks).Error; err != nil {
		return
	}
	for _, task := range tasks {
		if ctx.Err() != nil {
			return
		}
		_ = s.ProcessTask(ctx, task.OwnerID, task.ID)
	}
	if s.relist != nil {
		s.relist.processDue(ctx, accountID)
	}
}

// failTask 记录失败、指数退避下次执行时间和业务审计事件。
func (s *Delivery) failTask(db *gorm.DB, task model.DeliveryTask, err error) error {
	delay := time.Duration(1<<min(task.Attempts, 6)) * time.Second
	updateErr := db.Model(&task).Updates(map[string]any{
		"status": model.DeliveryTaskFailed, "last_error": err.Error(), "next_attempt_at": time.Now().Add(delay),
	}).Error
	s.events.Record(db.Statement.Context, model.BusinessEvent{OwnerID: task.OwnerID, AccountID: task.AccountID, OrderID: task.OrderID, Type: "delivery", Status: "failed", Summary: "自动发货失败", Error: err.Error()})
	return errors.Join(err, updateErr)
}

// deliveryResourceAAD 生成绑定规则和内容摘要的卡密加密附加数据。
func deliveryResourceAAD(ruleID uint, hash string) string {
	return fmt.Sprintf("xy:delivery-resource:%d:%s", ruleID, hash)
}

// deliveryTaskAAD 生成绑定账号和外部订单的发货负载附加数据。
func deliveryTaskAAD(accountID uint, orderID string) string {
	return fmt.Sprintf("xy:delivery-task:%d:%s", accountID, orderID)
}
