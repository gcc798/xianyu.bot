package agent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/response"
	"gorm.io/gorm"
)

type Session struct {
	db *gorm.DB
}

func NewSession(db *gorm.DB) *Session {
	return &Session{db: db}
}

func (s *Session) Create(ctx context.Context, ownerID uint, input xyRequest.CreateAgentSession) (model.AgentSession, error) {
	if ownerID == 0 {
		return model.AgentSession{}, errors.New("用户未登录")
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "新对话"
	}
	session := model.AgentSession{OwnerID: ownerID, Title: title}
	err := s.db.WithContext(ctx).Create(&session).Error
	return session, err
}

func (s *Session) List(ctx context.Context, ownerID uint, input xyRequest.AgentSessionSearch) ([]model.AgentSession, int64, error) {
	if ownerID == 0 {
		return nil, 0, errors.New("用户未登录")
	}
	db := s.db.WithContext(ctx).Model(&model.AgentSession{}).Where("owner_id = ?", ownerID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	sessions := make([]model.AgentSession, 0)
	err := db.Scopes(input.Paginate()).Order("last_message_at IS NULL, last_message_at DESC, id DESC").Find(&sessions).Error
	return sessions, total, err
}

func (s *Session) Get(ctx context.Context, ownerID, sessionID uint) (xyResponse.AgentSessionDetail, error) {
	if ownerID == 0 || sessionID == 0 {
		return xyResponse.AgentSessionDetail{}, errors.New("会话 ID 不能为空")
	}
	var session model.AgentSession
	if err := s.db.WithContext(ctx).Where("id = ? AND owner_id = ?", sessionID, ownerID).First(&session).Error; err != nil {
		return xyResponse.AgentSessionDetail{}, err
	}
	messages := make([]model.AgentMessage, 0)
	if err := s.db.WithContext(ctx).Where("session_id = ? AND owner_id = ?", sessionID, ownerID).Order("sequence").Find(&messages).Error; err != nil {
		return xyResponse.AgentSessionDetail{}, err
	}
	toolTraces := make([]xyResponse.AgentToolTrace, 0)
	if err := s.db.WithContext(ctx).Table("xy_agent_run_tools AS t").
		Select("t.id, t.run_id, t.call_id, t.tool_name, t.arguments, t.result, t.status, t.duration_millis, t.created_at").
		Joins("JOIN xy_agent_runs AS r ON r.id = t.run_id AND r.deleted_at IS NULL").
		Where("r.session_id = ? AND r.owner_id = ? AND t.deleted_at IS NULL", sessionID, ownerID).
		Order("t.id").Limit(200).Scan(&toolTraces).Error; err != nil {
		return xyResponse.AgentSessionDetail{}, err
	}
	approvals := make([]xyResponse.AgentApproval, 0)
	if err := s.db.WithContext(ctx).Table("xy_agent_run_tools AS t").
		Select("t.run_id, t.tool_name, t.created_at").
		Joins("JOIN xy_agent_runs AS r ON r.id = t.run_id AND r.deleted_at IS NULL").
		Where("r.session_id = ? AND r.owner_id = ? AND t.status = ? AND t.deleted_at IS NULL", sessionID, ownerID, "waiting_approval").
		Order("t.id").Scan(&approvals).Error; err != nil {
		return xyResponse.AgentSessionDetail{}, err
	}
	tasks := make([]xyResponse.AgentTask, 0)
	if err := s.db.WithContext(ctx).Table("xy_resource_tasks AS t").
		Select("t.id, t.agent_run_id AS run_id, t.state, t.stage, t.progress_percent, t.progress_done, t.progress_total, t.progress_unit").
		Joins("JOIN xy_agent_runs AS r ON r.id = t.agent_run_id AND r.deleted_at IS NULL").
		Where("r.session_id = ? AND r.owner_id = ? AND t.deleted_at IS NULL", sessionID, ownerID).
		Order("t.id").Limit(50).Scan(&tasks).Error; err != nil {
		return xyResponse.AgentSessionDetail{}, err
	}
	return xyResponse.AgentSessionDetail{AgentSession: session, Messages: messages, ToolTraces: toolTraces, PendingApprovals: approvals, Tasks: tasks}, nil
}

func (s *Session) Trajectory(ctx context.Context, ownerID, sessionID uint, input xyRequest.AgentTrajectorySearch) (xyResponse.AgentTrajectoryPage, error) {
	if ownerID == 0 || sessionID == 0 {
		return xyResponse.AgentTrajectoryPage{}, errors.New("会话 ID 不能为空")
	}
	var exists int64
	if err := s.db.WithContext(ctx).Model(&model.AgentSession{}).Where("id = ? AND owner_id = ?", sessionID, ownerID).Count(&exists).Error; err != nil {
		return xyResponse.AgentTrajectoryPage{}, err
	}
	if exists == 0 {
		return xyResponse.AgentTrajectoryPage{}, gorm.ErrRecordNotFound
	}
	limit := input.Limit
	if limit == 0 {
		limit = 200
	}
	events := make([]model.AgentEvent, 0, limit+1)
	query := s.db.WithContext(ctx).Where("owner_id = ? AND session_id = ? AND id > ?", ownerID, sessionID, input.AfterID)
	if input.RunID > 0 {
		query = query.Where("run_id = ?", input.RunID)
	}
	if eventType := strings.TrimSpace(input.Type); eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	if err := query.Order("id").Limit(limit + 1).Find(&events).Error; err != nil {
		return xyResponse.AgentTrajectoryPage{}, err
	}
	hasMore := len(events) > limit
	if hasMore {
		events = events[:limit]
	}
	items := make([]xyResponse.AgentTrajectoryEvent, 0, len(events))
	var nextAfterID uint64
	for _, event := range events {
		items = append(items, xyResponse.AgentTrajectoryEvent{
			ID: event.ID, RunID: event.RunID, TurnNo: event.TurnNo, StepNo: event.StepNo, Type: event.EventType,
			CallID: event.CallID, ParentCallID: event.ParentCallID, Data: json.RawMessage(event.Payload), CreatedAt: event.CreatedAt,
		})
		nextAfterID = event.ID
	}
	return xyResponse.AgentTrajectoryPage{Events: items, NextAfterID: nextAfterID, HasMore: hasMore}, nil
}

func (s *Session) ToolCall(ctx context.Context, ownerID, sessionID uint, callID string) (xyResponse.AgentToolCallDetail, error) {
	callID = strings.TrimSpace(callID)
	if ownerID == 0 || sessionID == 0 || callID == "" || len(callID) > 200 {
		return xyResponse.AgentToolCallDetail{}, errors.New("工具调用标识无效")
	}
	var call model.AgentRunTool
	if err := s.db.WithContext(ctx).Table("xy_agent_run_tools AS t").Select("t.*").
		Joins("JOIN xy_agent_runs AS r ON r.id = t.run_id AND r.deleted_at IS NULL").
		Where("r.owner_id = ? AND r.session_id = ? AND t.call_id = ? AND t.deleted_at IS NULL", ownerID, sessionID, callID).
		First(&call).Error; err != nil {
		return xyResponse.AgentToolCallDetail{}, err
	}
	return xyResponse.AgentToolCallDetail{
		ID: call.ID, RunID: call.RunID, CallID: call.CallID, MCPServerCode: call.MCPServerCode, ToolName: call.ToolName,
		Arguments: rawJSON(call.Arguments), Result: rawJSON(call.Result), Status: call.Status, DurationMillis: call.DurationMillis,
		ErrorCode: call.ErrorCode, ErrorMessage: call.ErrorMessage, StartedAt: call.StartedAt, CompletedAt: call.CompletedAt,
	}, nil
}

func rawJSON(value string) json.RawMessage {
	if json.Valid([]byte(value)) {
		return json.RawMessage(value)
	}
	return json.RawMessage("null")
}

func (s *Session) Delete(ctx context.Context, ownerID, sessionID uint) error {
	if ownerID == 0 || sessionID == 0 {
		return errors.New("会话 ID 不能为空")
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var session model.AgentSession
		if err := tx.Where("id = ? AND owner_id = ?", sessionID, ownerID).First(&session).Error; err != nil {
			return err
		}
		var running int64
		if err := tx.Model(&model.AgentRun{}).Where("session_id = ? AND owner_id = ? AND status = ?", sessionID, ownerID, "running").Count(&running).Error; err != nil {
			return err
		}
		if running > 0 {
			return errors.New("会话正在运行，暂不能删除")
		}
		var activeTasks int64
		if err := tx.Table("xy_resource_tasks AS t").
			Joins("JOIN xy_agent_runs AS r ON r.id = t.agent_run_id AND r.deleted_at IS NULL").
			Where("r.session_id = ? AND r.owner_id = ? AND t.state IN ? AND t.deleted_at IS NULL", sessionID, ownerID, []string{"queued", "running"}).
			Count(&activeTasks).Error; err != nil {
			return err
		}
		if activeTasks > 0 {
			return errors.New("会话仍有资源任务，暂不能删除")
		}
		runIDs := tx.Model(&model.AgentRun{}).Select("id").Where("session_id = ? AND owner_id = ?", sessionID, ownerID)
		if err := tx.Where("run_id IN (?)", runIDs).Delete(&model.AgentRunTool{}).Error; err != nil {
			return err
		}
		if err := tx.Where("session_id = ? AND owner_id = ?", sessionID, ownerID).Delete(&model.AgentEvent{}).Error; err != nil {
			return err
		}
		if err := tx.Where("session_id = ? AND owner_id = ?", sessionID, ownerID).Delete(&model.AgentRun{}).Error; err != nil {
			return err
		}
		if err := tx.Where("session_id = ? AND owner_id = ?", sessionID, ownerID).Delete(&model.AgentMessage{}).Error; err != nil {
			return err
		}
		return tx.Delete(&session).Error
	})
}
