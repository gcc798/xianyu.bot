package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gcc798/xianyu.bot/internal/xianyu/protocol"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// virtualProductFallbackReply 是虚拟商品未命中规则时的默认回复。
const virtualProductFallbackReply = "虚拟商品，下单自动网盘发货"

// textSender 描述消息服务需要的闲鱼文本发送能力。
type textSender interface {
	// SendText 向指定账号会话和接收方发送文本。
	SendText(context.Context, uint, uint, string, string, string) error
}

// replyResolver 描述消息服务需要的回复规则匹配能力。
type replyResolver interface {
	// Resolve 为用户商品和消息选择最高优先级规则。
	Resolve(context.Context, uint, uint, string) (*xyResponse.ReplyRule, error)
}

// orderRecorder 描述消息服务需要的订单消息入库能力。
type orderRecorder interface {
	// UpsertFromMessage 从闲鱼消息更新订单，非订单消息返回 nil。
	UpsertFromMessage(context.Context, model.Account, model.Message) (*model.Order, error)
}

// deliveryPlanner 描述订单产生后规划自动发货的能力。
type deliveryPlanner interface {
	// Plan 为符合条件的订单幂等创建发货任务。
	Plan(context.Context, model.Order) error
}

// PendingReply 保存内存中等待执行的会话回复。
type PendingReply struct {
	id     uint               // id 是持久化回复任务 ID。
	cancel context.CancelFunc // cancel 用于新消息或账号停止时取消等待。
}

// Message 管理闲鱼消息入库、人工回复和自动回复调度。
type Message struct {
	root     context.Context         // root 控制后台回复任务生命周期。
	db       *gorm.DB                // db 持久化消息、接管和回复任务。
	events   *BusinessEvent          // events 记录消息业务结果。
	sender   textSender              // sender 向闲鱼会话发送文本。
	rules    replyResolver           // rules 匹配商品回复规则。
	orders   orderRecorder           // orders 从消息提取并保存订单。
	delivery deliveryPlanner         // delivery 为消息订单规划发货。
	mu       sync.Mutex              // mu 保护 pending。
	pending  map[string]PendingReply // pending 按账号和会话保存等待中的回复。
}

// NewMessage 创建消息服务并初始化后台回复队列。
func NewMessage(ctx context.Context, db *gorm.DB, events *BusinessEvent, rules *ReplyRule, orders *Order, delivery *Delivery) *Message {
	return &Message{root: contextOrBackground(ctx), db: db, events: events, rules: rules, orders: orders, delivery: delivery, pending: make(map[string]PendingReply)}
}

// StartAccount 恢复账号未执行回复，并把重启时发送中的任务标记为未知。
func (s *Message) StartAccount(accountID uint) {
	var tasks []model.ReplyTask
	db := s.db
	if db == nil {
		return
	}
	_ = db.Model(&model.ReplyTask{}).Where("account_id = ? AND status = ?", accountID, model.ReplyTaskSending).
		Updates(map[string]any{"status": model.ReplyTaskUnknown, "error": "服务重启时发送结果未知"}).Error
	if db.Where("account_id = ? AND status = ?", accountID, model.ReplyTaskScheduled).Order("scheduled_at").Find(&tasks).Error != nil {
		return
	}
	for _, task := range tasks {
		s.launchReply(task)
	}
}

// StopAccount 取消指定账号所有仍在等待的自动回复。
func (s *Message) StopAccount(accountID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, task := range s.pending {
		if strings.HasPrefix(key, fmt.Sprintf("%d:", accountID)) {
			task.cancel()
			delete(s.pending, key)
		}
	}
}

// Handle 解析实时同步帧并处理其中的闲鱼消息。
func (s *Message) Handle(ctx context.Context, accountID uint, raw []byte) error {
	messages, err := protocol.ParseSyncEnvelope(raw)
	if err != nil || len(messages) == 0 {
		return err
	}
	return s.handleParsed(ctx, accountID, messages, false)
}

// HandleHistory 解析指定会话的历史响应，仅为最新且足够新的消息规划回复。
func (s *Message) HandleHistory(ctx context.Context, accountID uint, sessionID string, raw []byte) error {
	messages, err := protocol.ParseHistoryResponse(raw, sessionID)
	if err != nil || len(messages) == 0 {
		return err
	}
	return s.handleParsed(ctx, accountID, messages, true)
}

// handleParsed 幂等保存归一化消息，并触发订单、发货和自动回复流程。
func (s *Message) handleParsed(ctx context.Context, accountID uint, messages []protocol.ChatMessage, history bool) error {
	var account model.Account
	if err := s.db.WithContext(ctx).Select("id", "owner_id", "external_user_id").First(&account, accountID).Error; err != nil {
		return err
	}
	latestHistory := make(map[string]string)
	if history {
		latestTime := make(map[string]int64)
		for _, item := range messages {
			if item.MessageTime >= latestTime[item.SessionID] {
				latestTime[item.SessionID] = item.MessageTime
				latestHistory[item.SessionID] = item.PNMID
			}
		}
	}
	for _, item := range messages {
		direction := "inbound"
		if sameGoofishUser(item.SenderUserID, account.ExternalUserID) {
			direction = "outbound"
		}
		record := model.Message{
			OwnerID: account.OwnerID, AccountID: accountID, PNMID: item.PNMID, SessionID: item.SessionID,
			ContentType: item.ContentType, Content: item.Content, SenderUserName: item.SenderUserName,
			SenderUserID: item.SenderUserID, SenderAppVersion: item.SenderAppVersion,
			SenderPlatform: item.SenderPlatform, ReminderURL: item.ReminderURL,
			ExternalProductID: item.ProductID, ExternalOrderID: item.OrderID, MessageTime: item.MessageTime,
			Direction: direction, Source: "remote",
		}
		result := s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&record)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 1 {
			s.events.Record(ctx, model.BusinessEvent{OwnerID: account.OwnerID, AccountID: account.ID, Type: "message", Status: "success", Summary: "收到并保存闲鱼消息"})
			if direction == "outbound" {
				s.mergeLocalOutbound(ctx, record)
				continue
			}
			if s.orders != nil {
				orderRecord, err := s.orders.UpsertFromMessage(ctx, account, record)
				if err != nil {
					return err
				}
				if orderRecord != nil && s.delivery != nil {
					if err := s.delivery.Plan(ctx, *orderRecord); err != nil {
						return err
					}
				}
			}
			if history && (latestHistory[item.SessionID] != item.PNMID || item.MessageTime < time.Now().Add(-5*time.Minute).UnixMilli()) {
				continue
			}
			if err := s.scheduleAutoReply(ctx, account, record); err != nil {
				return err
			}
		}
	}
	return nil
}

// mergeLocalOutbound 用远端回执补全相邻的本地出站记录并删除重复消息。
func (s *Message) mergeLocalOutbound(ctx context.Context, remote model.Message) {
	if strings.TrimSpace(remote.Content) == "" {
		return
	}
	db := s.db.WithContext(ctx)
	var local model.Message
	err := db.Where("account_id = ? AND session_id = ? AND direction = ? AND source <> ? AND content = ? AND message_time BETWEEN ? AND ?",
		remote.AccountID, remote.SessionID, "outbound", "remote", remote.Content, remote.MessageTime-int64((2*time.Minute)/time.Millisecond), remote.MessageTime+int64((2*time.Minute)/time.Millisecond)).
		Order("id DESC").First(&local).Error
	if err != nil {
		return
	}
	_ = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Delete(&remote).Error; err != nil {
			return err
		}
		return tx.Model(&local).Updates(map[string]any{
			"pnm_id": remote.PNMID, "sender_user_id": remote.SenderUserID, "sender_user_name": remote.SenderUserName,
			"sender_app_version": remote.SenderAppVersion, "sender_platform": remote.SenderPlatform,
			"reminder_url": remote.ReminderURL, "message_time": remote.MessageTime, "source": "remote",
		}).Error
	})
}

// sameGoofishUser 忽略 @goofish 后缀比较两个非空用户标识。
func sameGoofishUser(left, right string) bool {
	normalize := func(value string) string {
		return strings.TrimSuffix(strings.TrimSpace(value), "@goofish")
	}
	return normalize(left) != "" && normalize(left) == normalize(right)
}

// List 分页查询 ownerID 名下消息并合并账号和人工接管状态。
func (s *Message) List(ctx context.Context, ownerID uint, input xyRequest.MessageSearch) ([]xyResponse.Message, int64, error) {
	if ownerID == 0 {
		return nil, 0, errors.New("用户未登录")
	}
	db := s.db.WithContext(ctx).Table("xy_messages AS m").Where("m.owner_id = ? AND m.deleted_at IS NULL", ownerID)
	if input.AccountID != 0 {
		db = db.Where("m.account_id = ?", input.AccountID)
	}
	if input.ProductID != 0 {
		db = db.Where(`EXISTS (
			SELECT 1 FROM xy_products AS p
			WHERE p.id = ? AND p.owner_id = m.owner_id AND p.account_id = m.account_id
				AND p.external_product_id = m.external_product_id AND p.deleted_at IS NULL
		)`, input.ProductID)
	}
	if input.Direction != "" {
		if input.Direction != "inbound" && input.Direction != "outbound" {
			return nil, 0, errors.New("消息方向无效")
		}
		db = db.Where("m.direction = ?", input.Direction)
	}
	if productID := strings.TrimSpace(input.ExternalProductID); productID != "" {
		db = db.Where("m.external_product_id = ?", productID)
	}
	if buyerName := strings.TrimSpace(input.BuyerName); buyerName != "" {
		db = db.Where("m.sender_user_name LIKE ?", "%"+buyerName+"%")
	}
	if content := strings.TrimSpace(input.Content); content != "" {
		db = db.Where("m.content LIKE ?", "%"+content+"%")
	}
	if input.StartTime > 0 && input.EndTime > 0 && input.EndTime < input.StartTime {
		return nil, 0, errors.New("结束时间不能早于开始时间")
	}
	if input.StartTime > 0 {
		db = db.Where("m.message_time >= ?", input.StartTime)
	}
	if input.EndTime > 0 {
		db = db.Where("m.message_time <= ?", input.EndTime)
	}
	if sessionID := strings.TrimSpace(input.SessionID); sessionID != "" {
		db = db.Where("m.session_id = ?", sessionID)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	list := make([]xyResponse.Message, 0)
	err := db.Select(`m.id, m.created_at, m.account_id, a.label AS account_label, m.pnm_id, m.session_id,
		m.content_type, m.content, m.sender_user_name, m.sender_user_id,
		m.external_product_id, m.external_order_id, m.message_time, m.direction, m.source,
		COALESCE(t.active, false) AS takeover_active`).
		Joins("JOIN xy_accounts AS a ON a.id = m.account_id AND a.deleted_at IS NULL").
		Joins("LEFT JOIN xy_conversation_takeovers AS t ON t.account_id = m.account_id AND t.session_id = m.session_id AND t.deleted_at IS NULL").
		Scopes(input.Paginate()).Order("m.message_time DESC, m.id DESC").Scan(&list).Error
	return list, total, err
}

// UpdateTakeover 依据归属消息启用或关闭对应会话的人工接管。
func (s *Message) UpdateTakeover(ctx context.Context, ownerID uint, input xyRequest.UpdateTakeover) error {
	item, err := s.ownedMessage(ctx, ownerID, input.MessageID)
	if err != nil {
		return err
	}
	if item.SessionID == "" {
		return errors.New("消息缺少会话 ID")
	}
	takeover := model.ConversationTakeover{
		OwnerID: ownerID, AccountID: item.AccountID, SessionID: item.SessionID, Active: input.Active,
	}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "account_id"}, {Name: "session_id"}},
		DoUpdates: clause.Assignments(map[string]any{"active": input.Active, "updated_at": time.Now()}),
	}).Create(&takeover).Error
}

// Reply 接管目标会话、发送人工回复并保存本地出站记录。
func (s *Message) Reply(ctx context.Context, ownerID uint, input xyRequest.SendMessage) error {
	text := strings.TrimSpace(input.Text)
	if input.MessageID == 0 || text == "" || len([]rune(text)) > 4000 {
		return errors.New("消息和回复内容不能为空，内容不能超过 4000 个字符")
	}
	if s.sender == nil {
		return errors.New("消息发送服务不可用")
	}
	item, err := s.ownedMessage(ctx, ownerID, input.MessageID)
	if err != nil {
		return err
	}
	if item.SessionID == "" || item.SenderUserID == "" {
		return errors.New("原消息缺少会话或接收方信息")
	}
	if err := s.UpdateTakeover(ctx, ownerID, xyRequest.UpdateTakeover{MessageID: item.ID, Active: true}); err != nil {
		return err
	}
	if err := s.sender.SendText(ctx, ownerID, item.AccountID, item.SessionID, item.SenderUserID, text); err != nil {
		return err
	}
	outgoing := model.Message{
		OwnerID: ownerID, AccountID: item.AccountID, PNMID: "local-" + uuid.NewString(), SessionID: item.SessionID,
		ContentType: 1, Content: text, SenderUserName: "我", ExternalProductID: item.ExternalProductID,
		ExternalOrderID: item.ExternalOrderID, MessageTime: time.Now().UnixMilli(), Direction: "outbound", Source: "manual",
	}
	if err := s.db.WithContext(ctx).Create(&outgoing).Error; err != nil {
		return errors.New("消息已发送，但本地记录保存失败")
	}
	return nil
}

// ownedMessage 查询 ownerID 名下的指定消息。
func (s *Message) ownedMessage(ctx context.Context, ownerID, messageID uint) (model.Message, error) {
	if ownerID == 0 || messageID == 0 {
		return model.Message{}, errors.New("消息 ID 不能为空")
	}
	var item model.Message
	err := s.db.WithContext(ctx).Where("id = ? AND owner_id = ?", messageID, ownerID).First(&item).Error
	return item, err
}

// scheduleAutoReply 校验商品设置、接管状态和规则后幂等调度回复。
func (s *Message) scheduleAutoReply(ctx context.Context, account model.Account, item model.Message) error {
	if s.sender == nil || s.rules == nil || item.ContentType != 1 || item.SessionID == "" ||
		item.ExternalProductID == "" || strings.TrimSpace(item.Content) == "" || item.SenderUserID == "" ||
		sameGoofishUser(item.SenderUserID, account.ExternalUserID) {
		return nil
	}
	var setting struct {
		ProductID         uint // ProductID 是消息关联的本地商品 ID。
		VirtualProduct    bool // VirtualProduct 决定是否允许默认回复。
		AutoReplyEnabled  bool // AutoReplyEnabled 表示商品已启用自动回复。
		ReplyDelaySeconds int  // ReplyDelaySeconds 是回复前等待秒数。
	}
	err := s.db.WithContext(ctx).Table("xy_products AS p").
		Select("p.id AS product_id, s.virtual_product, s.auto_reply_enabled, s.reply_delay_seconds").
		Joins("JOIN xy_product_settings AS s ON s.product_id = p.id AND s.deleted_at IS NULL").
		Where("p.owner_id = ? AND p.account_id = ? AND p.external_product_id = ? AND p.deleted_at IS NULL",
			account.OwnerID, account.ID, item.ExternalProductID).
		Take(&setting).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || !setting.AutoReplyEnabled {
		return nil
	}
	if err != nil {
		return err
	}
	var takeoverCount int64
	if err := s.db.WithContext(ctx).Model(&model.ConversationTakeover{}).
		Where("owner_id = ? AND account_id = ? AND session_id = ? AND active = ?", account.OwnerID, account.ID, item.SessionID, true).
		Count(&takeoverCount).Error; err != nil || takeoverCount > 0 {
		return err
	}
	rule, err := s.rules.Resolve(ctx, account.OwnerID, setting.ProductID, item.Content)
	if err != nil {
		return err
	}
	if rule == nil && setting.VirtualProduct {
		rule = &xyResponse.ReplyRule{Content: virtualProductFallbackReply}
	}
	if rule == nil {
		return nil
	}
	now := time.Now()
	task := model.ReplyTask{
		OwnerID: account.OwnerID, AccountID: account.ID, MessageID: item.ID, ProductID: setting.ProductID,
		ExternalProductID: item.ExternalProductID,
		RuleID:            rule.ID, SessionID: item.SessionID, RecipientID: item.SenderUserID, ReplyText: rule.Content,
		Status: model.ReplyTaskScheduled, ScheduledAt: now.Add(time.Duration(setting.ReplyDelaySeconds) * time.Second),
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		completedAt := now
		if err := tx.Model(&model.ReplyTask{}).
			Where("account_id = ? AND session_id = ? AND status = ?", account.ID, item.SessionID, model.ReplyTaskScheduled).
			Updates(map[string]any{"status": model.ReplyTaskCancelled, "completed_at": completedAt, "error": "被同会话的新消息合并"}).Error; err != nil {
			return err
		}
		if err := tx.Create(&task).Error; err != nil {
			return err
		}
		return nil
	})
	if err == nil {
		s.launchReply(task)
	}
	return err
}

// launchReply 按账号和会话合并等待任务，并在计划时间启动发送。
func (s *Message) launchReply(task model.ReplyTask) {
	key := fmt.Sprintf("%d:%s", task.AccountID, task.SessionID)
	ctx, cancel := context.WithCancel(contextOrBackground(s.root))
	s.mu.Lock()
	if s.pending == nil {
		s.pending = make(map[string]PendingReply)
	}
	if previous := s.pending[key]; previous.cancel != nil {
		previous.cancel()
	}
	if len(s.pending) >= 256 {
		s.mu.Unlock()
		cancel()
		_ = s.finishReplyTask(task.ID, model.ReplyTaskFailed, "待回复队列已满")
		return
	}
	s.pending[key] = PendingReply{id: task.ID, cancel: cancel}
	s.mu.Unlock()
	go func() {
		delay := time.Until(task.ScheduledAt)
		if delay > 0 {
			timer := time.NewTimer(delay)
			defer timer.Stop()
			select {
			case <-timer.C:
			case <-ctx.Done():
				return
			}
		}
		s.runReplyTask(ctx, task)
		s.mu.Lock()
		if current := s.pending[key]; current.id == task.ID {
			delete(s.pending, key)
		}
		s.mu.Unlock()
	}()
}

// runReplyTask 重新校验接管与商品设置后发送并落库回复结果。
func (s *Message) runReplyTask(ctx context.Context, task model.ReplyTask) {
	db := s.db.WithContext(ctx)
	result := db.Model(&model.ReplyTask{}).Where("id = ? AND status = ?", task.ID, model.ReplyTaskScheduled).
		Update("status", model.ReplyTaskSending)
	if result.Error != nil || result.RowsAffected != 1 {
		return
	}
	var takeoverCount int64
	err := db.Model(&model.ConversationTakeover{}).
		Where("account_id = ? AND session_id = ? AND active = ?", task.AccountID, task.SessionID, true).
		Count(&takeoverCount).Error
	if err != nil {
		_ = s.finishReplyTask(task.ID, model.ReplyTaskFailed, "检查人工接管状态失败")
		return
	}
	if takeoverCount > 0 {
		_ = s.finishReplyTask(task.ID, model.ReplyTaskCancelled, "会话已人工接管")
		return
	}
	var setting model.ProductSetting
	if err := db.Where("product_id = ? AND auto_reply_enabled = ?", task.ProductID, true).First(&setting).Error; err != nil {
		_ = s.finishReplyTask(task.ID, model.ReplyTaskCancelled, "商品自动回复已关闭")
		return
	}
	if task.RuleID == 0 && !setting.VirtualProduct {
		_ = s.finishReplyTask(task.ID, model.ReplyTaskCancelled, "商品已关闭虚拟商品设置")
		return
	}
	if err := s.sender.SendText(ctx, task.OwnerID, task.AccountID, task.SessionID, task.RecipientID, task.ReplyText); err != nil {
		status := model.ReplyTaskFailed
		if errors.Is(err, errSendUnknown) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			status = model.ReplyTaskUnknown
		}
		_ = s.finishReplyTask(task.ID, status, err.Error())
		return
	}
	now := time.Now()
	_ = db.Transaction(func(tx *gorm.DB) error {
		outgoing := model.Message{
			OwnerID: task.OwnerID, AccountID: task.AccountID, PNMID: "local-" + uuid.NewString(), SessionID: task.SessionID,
			ContentType: 1, Content: task.ReplyText, SenderUserName: "自动回复", MessageTime: now.UnixMilli(),
			ExternalProductID: task.ExternalProductID, Direction: "outbound", Source: "keyword",
		}
		if err := tx.Create(&outgoing).Error; err != nil {
			return err
		}
		return tx.Model(&model.ReplyTask{}).Where("id = ?", task.ID).
			Updates(map[string]any{"status": model.ReplyTaskSent, "completed_at": now, "error": ""}).Error
	})
}

// finishReplyTask 原子写入回复任务的终态、完成时间和说明。
func (s *Message) finishReplyTask(id uint, status, message string) error {
	now := time.Now()
	return s.db.Model(&model.ReplyTask{}).Where("id = ?", id).
		Updates(map[string]any{"status": status, "completed_at": now, "error": message}).Error
}
