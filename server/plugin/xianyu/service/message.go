package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xianyu/model"
	xianyuRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/xianyu/model/request"
	xianyuResponse "github.com/flipped-aurora/gin-vue-admin/server/plugin/xianyu/model/response"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xianyu/protocol"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type textSender interface {
	SendText(context.Context, uint, uint, string, string, string) error
}

type replyResolver interface {
	Resolve(context.Context, uint, uint, string) (*xianyuResponse.ReplyRule, error)
}

type orderRecorder interface {
	UpsertFromMessage(context.Context, model.Account, model.Message) (*model.Order, error)
}

type deliveryPlanner interface {
	Plan(context.Context, model.Order) error
}

type pendingReply struct {
	id     uint
	cancel context.CancelFunc
}

type message struct {
	sender   textSender
	rules    replyResolver
	orders   orderRecorder
	delivery deliveryPlanner
	mu       sync.Mutex
	pending  map[string]pendingReply
}

func (s *message) StartAccount(accountID uint) {
	var tasks []model.ReplyTask
	db := global.GVA_DB
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

func (s *message) StopAccount(accountID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, task := range s.pending {
		if strings.HasPrefix(key, fmt.Sprintf("%d:", accountID)) {
			task.cancel()
			delete(s.pending, key)
		}
	}
}

func (s *message) Handle(ctx context.Context, accountID uint, raw []byte) error {
	messages, err := protocol.ParseSyncEnvelope(raw)
	if err != nil || len(messages) == 0 {
		return err
	}
	return s.handleParsed(ctx, accountID, messages, false)
}

func (s *message) HandleHistory(ctx context.Context, accountID uint, sessionID string, raw []byte) error {
	messages, err := protocol.ParseHistoryResponse(raw, sessionID)
	if err != nil || len(messages) == 0 {
		return err
	}
	return s.handleParsed(ctx, accountID, messages, true)
}

func (s *message) handleParsed(ctx context.Context, accountID uint, messages []protocol.ChatMessage, history bool) error {
	var account model.Account
	if err := global.GVA_DB.WithContext(ctx).Select("id", "owner_id", "external_user_id").First(&account, accountID).Error; err != nil {
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
		result := global.GVA_DB.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&record)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 1 {
			recordBusinessEvent(ctx, model.BusinessEvent{OwnerID: account.OwnerID, AccountID: account.ID, Type: "message", Status: "success", Summary: "收到并保存闲鱼消息"})
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

func (s *message) mergeLocalOutbound(ctx context.Context, remote model.Message) {
	if strings.TrimSpace(remote.Content) == "" {
		return
	}
	db := global.GVA_DB.WithContext(ctx)
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

func sameGoofishUser(left, right string) bool {
	normalize := func(value string) string {
		return strings.TrimSuffix(strings.TrimSpace(value), "@goofish")
	}
	return normalize(left) != "" && normalize(left) == normalize(right)
}

func (s *message) List(ctx context.Context, ownerID uint, input xianyuRequest.MessageSearch) ([]xianyuResponse.Message, int64, error) {
	if ownerID == 0 {
		return nil, 0, errors.New("用户未登录")
	}
	db := global.GVA_DB.WithContext(ctx).Table("xy_messages AS m").Where("m.owner_id = ? AND m.deleted_at IS NULL", ownerID)
	if input.AccountID != 0 {
		db = db.Where("m.account_id = ?", input.AccountID)
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
	list := make([]xianyuResponse.Message, 0)
	err := db.Select(`m.id, m.created_at, m.account_id, a.label AS account_label, m.pnm_id, m.session_id,
		m.content_type, m.content, m.sender_user_name, m.sender_user_id,
		m.external_product_id, m.external_order_id, m.message_time, m.direction, m.source,
		COALESCE(t.active, false) AS takeover_active`).
		Joins("JOIN xy_accounts AS a ON a.id = m.account_id AND a.deleted_at IS NULL").
		Joins("LEFT JOIN xy_conversation_takeovers AS t ON t.account_id = m.account_id AND t.session_id = m.session_id AND t.deleted_at IS NULL").
		Scopes(input.Paginate()).Order("m.message_time DESC, m.id DESC").Scan(&list).Error
	return list, total, err
}

func (s *message) UpdateTakeover(ctx context.Context, ownerID uint, input xianyuRequest.UpdateTakeover) error {
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
	return global.GVA_DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "account_id"}, {Name: "session_id"}},
		DoUpdates: clause.Assignments(map[string]any{"active": input.Active, "updated_at": time.Now()}),
	}).Create(&takeover).Error
}

func (s *message) Reply(ctx context.Context, ownerID uint, input xianyuRequest.SendMessage) error {
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
	if err := s.UpdateTakeover(ctx, ownerID, xianyuRequest.UpdateTakeover{MessageID: item.ID, Active: true}); err != nil {
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
	if err := global.GVA_DB.WithContext(ctx).Create(&outgoing).Error; err != nil {
		return errors.New("消息已发送，但本地记录保存失败")
	}
	return nil
}

func (s *message) ownedMessage(ctx context.Context, ownerID, messageID uint) (model.Message, error) {
	if ownerID == 0 || messageID == 0 {
		return model.Message{}, errors.New("消息 ID 不能为空")
	}
	var item model.Message
	err := global.GVA_DB.WithContext(ctx).Where("id = ? AND owner_id = ?", messageID, ownerID).First(&item).Error
	return item, err
}

func (s *message) scheduleAutoReply(ctx context.Context, account model.Account, item model.Message) error {
	if s.sender == nil || s.rules == nil || item.ContentType != 1 || item.SessionID == "" ||
		item.ExternalProductID == "" || strings.TrimSpace(item.Content) == "" || item.SenderUserID == "" ||
		sameGoofishUser(item.SenderUserID, account.ExternalUserID) {
		return nil
	}
	var setting struct {
		ProductID         uint
		AutoReplyEnabled  bool
		ReplyDelaySeconds int
	}
	err := global.GVA_DB.WithContext(ctx).Table("xy_products AS p").
		Select("p.id AS product_id, s.auto_reply_enabled, s.reply_delay_seconds").
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
	if err := global.GVA_DB.WithContext(ctx).Model(&model.ConversationTakeover{}).
		Where("owner_id = ? AND account_id = ? AND session_id = ? AND active = ?", account.OwnerID, account.ID, item.SessionID, true).
		Count(&takeoverCount).Error; err != nil || takeoverCount > 0 {
		return err
	}
	rule, err := s.rules.Resolve(ctx, account.OwnerID, setting.ProductID, item.Content)
	if err != nil || rule == nil {
		return err
	}
	now := time.Now()
	task := model.ReplyTask{
		OwnerID: account.OwnerID, AccountID: account.ID, MessageID: item.ID, ProductID: setting.ProductID,
		ExternalProductID: item.ExternalProductID,
		RuleID:            rule.ID, SessionID: item.SessionID, RecipientID: item.SenderUserID, ReplyText: rule.Content,
		Status: model.ReplyTaskScheduled, ScheduledAt: now.Add(time.Duration(setting.ReplyDelaySeconds) * time.Second),
	}
	err = global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

func (s *message) launchReply(task model.ReplyTask) {
	key := fmt.Sprintf("%d:%s", task.AccountID, task.SessionID)
	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	if s.pending == nil {
		s.pending = make(map[string]pendingReply)
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
	s.pending[key] = pendingReply{id: task.ID, cancel: cancel}
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

func (s *message) runReplyTask(ctx context.Context, task model.ReplyTask) {
	db := global.GVA_DB.WithContext(ctx)
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
	if err := s.sender.SendText(ctx, task.OwnerID, task.AccountID, task.SessionID, task.RecipientID, task.ReplyText); err != nil {
		status := model.ReplyTaskFailed
		if strings.Contains(err.Error(), "结果未知") || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
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

func (s *message) finishReplyTask(id uint, status, message string) error {
	now := time.Now()
	return global.GVA_DB.Model(&model.ReplyTask{}).Where("id = ?", id).
		Updates(map[string]any{"status": status, "completed_at": now, "error": message}).Error
}
