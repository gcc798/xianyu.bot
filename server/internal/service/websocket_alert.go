package service

import (
	"fmt"
	"html"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/gcc798/xianyu.bot/internal/xianyu/connection"
	"go.uber.org/zap"
)

type connectionAlertWindow struct {
	startedAt        time.Time
	unavailable      bool
	unavailableCount uint
	alerted          bool
	lastFailure      connection.Snapshot
	notificationDone chan struct{}
}

func (s *WebSocketService) connectionStateChanged(snapshot connection.Snapshot) {
	fields := []zap.Field{
		zap.Uint("accountID", snapshot.AccountID), zap.String("state", string(snapshot.State)),
		zap.Uint64("reconnects", snapshot.Reconnects), zap.String("lastError", snapshot.LastError),
	}
	if snapshot.State == connection.StateConnected || snapshot.State == connection.StateConnecting {
		zap.L().Info("闲鱼消息连接状态变化", fields...)
	} else {
		zap.L().Warn("闲鱼消息连接状态变化", fields...)
	}

	s.trackConnectionAlertWindow(snapshot)
	if snapshot.State == connection.StateFailed {
		s.stopAccountWorkers(snapshot.AccountID)
	}
}

func (s *WebSocketService) trackConnectionAlertWindow(snapshot connection.Snapshot) {
	if snapshot.State != connection.StateConnected && snapshot.State != connection.StateReconnecting && snapshot.State != connection.StateFailed {
		return
	}

	s.alertMu.Lock()
	window := s.alertWindows[snapshot.AccountID]
	if snapshot.State == connection.StateConnected {
		if window == nil {
			s.alertMu.Unlock()
			return
		}
		window.unavailable = false
		if !window.alerted {
			s.alertMu.Unlock()
			return
		}
		delete(s.alertWindows, snapshot.AccountID)
		notificationDone := window.notificationDone
		s.alertMu.Unlock()
		go func() {
			<-notificationDone
			s.sendConnectionNotification(snapshot, true, window.startedAt, window.unavailableCount)
		}()
		return
	}

	if window == nil {
		window = &connectionAlertWindow{startedAt: time.Now()}
		s.alertWindows[snapshot.AccountID] = window
		duration := s.alertWindowDuration
		if duration <= 0 {
			duration = time.Hour
		}
		go func() {
			timer := time.NewTimer(duration)
			defer timer.Stop()
			select {
			case <-contextOrBackground(s.root).Done():
			case <-timer.C:
				s.finishConnectionAlertWindow(snapshot.AccountID, window)
			}
		}()
	}
	if !window.unavailable {
		window.unavailableCount++
	}
	window.unavailable = true
	window.lastFailure = snapshot
	s.alertMu.Unlock()
}

func (s *WebSocketService) finishConnectionAlertWindow(accountID uint, target *connectionAlertWindow) {
	s.alertMu.Lock()
	window := s.alertWindows[accountID]
	if window != target || window.alerted {
		s.alertMu.Unlock()
		return
	}
	if !window.unavailable {
		delete(s.alertWindows, accountID)
		s.alertMu.Unlock()
		s.recordSuppressedConnectionAlert(accountID, window)
		return
	}
	window.alerted = true
	window.notificationDone = make(chan struct{})
	snapshot := window.lastFailure
	s.alertMu.Unlock()
	s.sendConnectionNotification(snapshot, false, window.startedAt, window.unavailableCount)
	close(window.notificationDone)
}

func (s *WebSocketService) recordSuppressedConnectionAlert(accountID uint, window *connectionAlertWindow) {
	var account model.Account
	if err := s.db.Where("id = ?", accountID).First(&account).Error; err != nil {
		zap.L().Error("记录闲鱼消息连接抖动失败", zap.Uint("accountID", accountID), zap.Error(err))
		return
	}
	s.events.Record(contextOrBackground(s.root), model.BusinessEvent{
		OwnerID: account.OwnerID, AccountID: account.ID, Type: "websocket_alert", Status: "success",
		Summary: fmt.Sprintf("一小时连接抖动窗口结束，当前连接正常（不可用 %d 次，未发送告警）", window.unavailableCount),
		Error:   window.lastFailure.LastError,
	})
}

func (s *WebSocketService) cancelConnectionAlertWindow(accountID uint) {
	s.alertMu.Lock()
	delete(s.alertWindows, accountID)
	s.alertMu.Unlock()
}

func (s *WebSocketService) sendConnectionNotification(snapshot connection.Snapshot, recovered bool, startedAt time.Time, unavailableCount uint) {
	var account model.Account
	if err := s.db.Where("id = ?", snapshot.AccountID).First(&account).Error; err != nil {
		zap.L().Error("读取闲鱼消息告警账号失败", zap.Uint("accountID", snapshot.AccountID), zap.Error(err))
		return
	}
	subject := "[闲鱼助手告警] 消息连接不可用"
	body := fmt.Sprintf("账号：%s（ID: %d）<br>状态：%s<br>原因：%s<br>窗口开始：%s<br>一小时内不可用次数：%d<br>时间：%s",
		html.EscapeString(account.Label), account.ID, html.EscapeString(string(snapshot.State)),
		html.EscapeString(snapshot.LastError), startedAt.Format(time.DateTime), unavailableCount, time.Now().Format(time.DateTime))
	summary := "已发送闲鱼消息连接故障邮件"
	if recovered {
		subject = "[闲鱼助手恢复] 消息连接已恢复"
		body = fmt.Sprintf("账号：%s（ID: %d）<br>状态：已恢复正常<br>故障窗口开始：%s<br>一小时内不可用次数：%d<br>时间：%s",
			html.EscapeString(account.Label), account.ID, startedAt.Format(time.DateTime), unavailableCount, time.Now().Format(time.DateTime))
		summary = "已发送闲鱼消息连接恢复邮件"
	} else if snapshot.ActionURL != "" {
		body += "<br>处理方式：请在管理后台点击“重新扫码验证”，并在闲鱼 App 完成安全验证。"
	}
	sender := s.sendNotificationEmail
	if sender == nil {
		return
	}
	if err := sender("force_gc@163.com", subject, body); err != nil {
		s.events.Record(contextOrBackground(s.root), model.BusinessEvent{
			OwnerID: account.OwnerID, AccountID: account.ID, Type: "websocket_alert", Status: "failed",
			Summary: summary + "失败", Error: err.Error(),
		})
		zap.L().Error(summary+"失败", zap.Uint("accountID", snapshot.AccountID), zap.Error(err))
		return
	}
	event := model.BusinessEvent{
		OwnerID: account.OwnerID, AccountID: account.ID, Type: "websocket_alert", Status: "success",
		Summary: fmt.Sprintf("%s（一小时内不可用 %d 次）", summary, unavailableCount),
	}
	if !recovered {
		event.Error = snapshot.LastError
	}
	s.events.Record(contextOrBackground(s.root), event)
	zap.L().Info(summary, zap.Uint("accountID", snapshot.AccountID), zap.String("recipient", "force_gc@163.com"))
}
