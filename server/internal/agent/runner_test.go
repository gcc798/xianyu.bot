package agent

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Run 提供测试辅助能力。
func (s *Runner) Run(ctx context.Context, ownerID, sessionID uint, content, idempotencyKey string, directives ...xyRequest.AgentDirective) (RunResult, error) {
	prepared, existing, err := s.prepare(ctx, ownerID, sessionID, content, idempotencyKey, directives)
	if err != nil {
		return RunResult{}, err
	}
	if existing != nil {
		return *existing, nil
	}
	s.publish(prepared.run.SessionID, "run.started", map[string]any{"runId": prepared.run.ID, "model": prepared.run.ModelName})
	return s.execute(ctx, prepared)
}

// TestAgentRunnerPersistsConversationAndDeduplicatesRequest 验证对应测试场景。
func TestAgentRunnerPersistsConversationAndDeduplicatesRequest(t *testing.T) {
	db := agentRunnerTestDB(t)
	session := model.AgentSession{OwnerID: 11, Title: "新对话"}
	require.NoError(t, db.Create(&session).Error)
	chatModel := &stubAgentModel{answer: "这是最终回答"}
	runner := &Runner{db: db, registry: &MCPRegistry{db: db}, newModel: func(context.Context, model.AIModel, model.AgentConfig) (einomodel.BaseChatModel, error) {
		return chatModel, nil
	}}

	result, err := runner.Run(t.Context(), 11, session.ID, "帮我找 Go 资料", "request-1")
	require.NoError(t, err)
	require.Equal(t, "completed", result.Run.Status)
	require.NotNil(t, result.Assistant)
	require.Equal(t, "这是最终回答", result.Assistant.Content)
	require.EqualValues(t, 1, chatModel.calls.Load())
	require.NotContains(t, result.Run.ConfigSnapshot, "secret-key")
	var renamed model.AgentSession
	require.NoError(t, db.First(&renamed, session.ID).Error)
	require.Equal(t, "帮我找 Go 资料", renamed.Title)

	repeated, err := runner.Run(t.Context(), 11, session.ID, "帮我找 Go 资料", "request-1")
	require.NoError(t, err)
	require.Equal(t, result.Run.ID, repeated.Run.ID)
	require.Equal(t, result.Assistant.ID, repeated.Assistant.ID)
	require.EqualValues(t, 1, chatModel.calls.Load())
	_, err = runner.Run(t.Context(), 11, session.ID, "不同内容", "request-1")
	require.ErrorContains(t, err, "其他消息")

	var messages []model.AgentMessage
	require.NoError(t, db.Where("session_id = ?", session.ID).Order("sequence").Find(&messages).Error)
	require.Len(t, messages, 2)
	require.Equal(t, "user", messages[0].Role)
	require.Equal(t, "assistant", messages[1].Role)
	_, err = runner.Run(t.Context(), 22, session.ID, "越权消息", "request-owner-22")
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

// TestAgentRunnerRecordsModelFailure 验证对应测试场景。
func TestAgentRunnerRecordsModelFailure(t *testing.T) {
	db := agentRunnerTestDB(t)
	session := model.AgentSession{OwnerID: 11, Title: "失败会话"}
	require.NoError(t, db.Create(&session).Error)
	runner := &Runner{db: db, registry: &MCPRegistry{db: db}, events: NewEventHub(db), newModel: func(context.Context, model.AIModel, model.AgentConfig) (einomodel.BaseChatModel, error) {
		return &stubAgentModel{err: errors.New("upstream unavailable")}, nil
	}}

	result, err := runner.Run(t.Context(), 11, session.ID, "测试失败", "request-failed")
	require.ErrorContains(t, err, "upstream unavailable")
	require.NotZero(t, result.Run.ID)
	var stored model.AgentRun
	require.NoError(t, db.First(&stored, result.Run.ID).Error)
	require.Equal(t, "failed", stored.Status)
	require.Equal(t, "failed", stored.EndReason)
	require.Equal(t, 1, stored.ModelCalls)
	var failedEvents int64
	require.NoError(t, db.Model(&model.AgentEvent{}).Where("session_id = ? AND event_type = ?", session.ID, "model.attempt.failed").Count(&failedEvents).Error)
	require.EqualValues(t, 1, failedEvents)
}

// TestAgentInstructionContainsIdentityWithoutRuntimeTime 验证基础提示词不混入动态时间。
func TestAgentInstructionContainsIdentityWithoutRuntimeTime(t *testing.T) {
	require.Contains(t, xyAgentInstruction, "你是闲鱼自动化助手，帮助用户完成资源检索、资源下载和闲鱼运营工作")
	require.NotContains(t, xyAgentInstruction, "当前时间：")
	require.Contains(t, xyOperationsInstruction, "账号配置、二维码登录、Cookie、Token")
	require.Contains(t, xyOperationsInstruction, "approval_required 只代表等待审批")
	require.Contains(t, xyOperationsInstruction, "unknown 表示平台结果未确认，禁止重试")
	require.Contains(t, xyOperationsInstruction, "不得在对话中索取、接收或复述卡密")
}

// TestAgentFailureMessageExplainsInvalidAPIKey 验证无效凭据不会只显示运行失败。
func TestAgentFailureMessageExplainsInvalidAPIKey(t *testing.T) {
	message := agentFailureMessage(errors.New(`status code: 401, body: {"code":"INVALID_API_KEY"}`))
	require.Equal(t, "模型 API Key 无效，请在系统设置中更新模型凭据", message)
}

// TestAgentRunnerFallsBackByPriority 验证高优先级模型失败后自动切换并记录轨迹。
func TestAgentRunnerFallsBackByPriority(t *testing.T) {
	db := agentRunnerTestDB(t)
	backup := model.AIModel{
		Name: "备用模型", Provider: "openai", BaseURL: "https://example.com/v1", APIKey: "backup-key",
		Model: "backup", Enabled: true, Priority: 2, SupportsTools: true, SupportsStreaming: true,
		ContextWindow: 128000, MaxOutputTokens: 2000, RequestTimeoutSeconds: 60,
	}
	require.NoError(t, db.Create(&backup).Error)
	session := model.AgentSession{OwnerID: 11, Title: "降级会话"}
	require.NoError(t, db.Create(&session).Error)
	runner := &Runner{db: db, registry: &MCPRegistry{db: db}, events: NewEventHub(db), newModel: func(_ context.Context, item model.AIModel, _ model.AgentConfig) (einomodel.BaseChatModel, error) {
		if item.Name == "测试模型" {
			return &stubAgentModel{err: errors.New("primary unavailable")}, nil
		}
		return &stubAgentModel{answer: "备用模型回答"}, nil
	}}

	result, err := runner.Run(t.Context(), 11, session.ID, "测试降级", "request-fallback")
	require.NoError(t, err)
	require.Equal(t, "备用模型回答", result.Assistant.Content)
	require.Equal(t, backup.ID, result.Run.ModelID)
	require.Equal(t, "备用模型", result.Run.ModelName)
	var fallbackEvents int64
	require.NoError(t, db.Model(&model.AgentEvent{}).Where("session_id = ? AND event_type = ?", session.ID, "model.fallback").Count(&fallbackEvents).Error)
	require.EqualValues(t, 1, fallbackEvents)
}

// TestAgentRunnerSubmitPublishesStreamingEvents 验证对应测试场景。
func TestAgentRunnerSubmitPublishesStreamingEvents(t *testing.T) {
	db := agentRunnerTestDB(t)
	session := model.AgentSession{OwnerID: 11, Title: "异步会话"}
	require.NoError(t, db.Create(&session).Error)
	events := NewEventHub(db)
	runner := &Runner{db: db, registry: &MCPRegistry{db: db}, events: events, newModel: func(context.Context, model.AIModel, model.AgentConfig) (einomodel.BaseChatModel, error) {
		return &stubAgentModel{answer: "流式回答"}, nil
	}}
	_, stream, cancel, err := events.Subscribe(t.Context(), session.ID, 0)
	require.NoError(t, err)
	defer cancel()

	accepted, err := runner.Submit(t.Context(), 11, session.ID, "异步问题", "request-async")
	require.NoError(t, err)
	require.Equal(t, "running", accepted.Run.Status)

	types := make([]string, 0, 12)
	deadline := time.After(3 * time.Second)
	for {
		select {
		case event := <-stream:
			types = append(types, event.Type)
			if event.Type == "run.completed" {
				require.Contains(t, types, "request.snapshot")
				require.Contains(t, types, "model.attempt.completed")
				require.Contains(t, types, "assistant.delta")
				require.Contains(t, types, "assistant.message")
				var stored model.AgentRun
				require.NoError(t, db.First(&stored, accepted.Run.ID).Error)
				require.Equal(t, "completed", stored.Status)
				return
			}
		case <-deadline:
			t.Fatalf("timed out waiting for events: %v", types)
		}
	}
}

// TestAgentRunnerPersistsAuditableModelSnapshot 验证对应测试场景。
func TestAgentRunnerPersistsAuditableModelSnapshot(t *testing.T) {
	db := agentRunnerTestDB(t)
	session := model.AgentSession{OwnerID: 11, Title: "模型轨迹"}
	require.NoError(t, db.Create(&session).Error)
	runner := &Runner{db: db, registry: &MCPRegistry{db: db}, events: NewEventHub(db), newModel: func(context.Context, model.AIModel, model.AgentConfig) (einomodel.BaseChatModel, error) {
		return &stubAgentModel{answer: "轨迹回答"}, nil
	}}

	result, err := runner.Run(t.Context(), 11, session.ID, "记录实际模型输入", "request-trace")
	require.NoError(t, err)
	var events []model.AgentEvent
	require.NoError(t, db.Where("session_id = ?", session.ID).Order("id").Find(&events).Error)
	types := make([]string, 0, len(events))
	for _, event := range events {
		types = append(types, event.EventType)
		require.NotContains(t, event.Payload, "secret-key")
	}
	require.Contains(t, types, "system.prompt")
	require.Contains(t, types, "request.snapshot")
	require.Contains(t, types, "model.attempt.completed")
	require.Contains(t, types, "assistant.message")
	require.Contains(t, types, "turn.completed")
	require.Equal(t, result.Run.ID, *events[len(events)-1].RunID)
}

// TestAgentRunnerCancelsOwnedRun 验证对应测试场景。
func TestAgentRunnerCancelsOwnedRun(t *testing.T) {
	db := agentRunnerTestDB(t)
	session := model.AgentSession{OwnerID: 11, Title: "取消会话"}
	require.NoError(t, db.Create(&session).Error)
	events := NewEventHub(db)
	runner := &Runner{db: db, registry: &MCPRegistry{db: db}, events: events, newModel: func(context.Context, model.AIModel, model.AgentConfig) (einomodel.BaseChatModel, error) {
		return cancelAwareAgentModel{}, nil
	}}
	_, stream, unsubscribe, err := events.Subscribe(t.Context(), session.ID, 0)
	require.NoError(t, err)
	defer unsubscribe()

	accepted, err := runner.Submit(t.Context(), 11, session.ID, "等待取消", "request-cancel")
	require.NoError(t, err)
	require.Error(t, runner.Cancel(t.Context(), 22, session.ID, accepted.Run.ID))
	require.NoError(t, runner.Cancel(t.Context(), 11, session.ID, accepted.Run.ID))

	deadline := time.After(3 * time.Second)
	for {
		select {
		case event := <-stream:
			if event.Type != "run.cancelled" {
				continue
			}
			var stored model.AgentRun
			require.NoError(t, db.First(&stored, accepted.Run.ID).Error)
			require.Equal(t, "cancelled", stored.Status)
			require.Equal(t, "cancelled", stored.EndReason)
			return
		case <-deadline:
			t.Fatal("timed out waiting for cancellation")
		}
	}
}

// TestValidateAgentDirectives 验证对应测试场景。
func TestValidateAgentDirectives(t *testing.T) {
	directives := []xyRequest.AgentDirective{
		{Type: " skill ", Name: " resource-discovery "},
		{Type: "tool", Name: "message_search"},
	}
	require.NoError(t, validateAgentDirectives(directives))
	require.Equal(t, "skill", directives[0].Type)
	require.Equal(t, "resource-discovery", directives[0].Name)
	require.ErrorContains(t, validateAgentDirectives([]xyRequest.AgentDirective{{Type: "skill", Name: "unknown"}}), "Skill 不存在")
	require.ErrorContains(t, validateAgentDirectives([]xyRequest.AgentDirective{{Type: "tool", Name: "bad tool"}}), "Tool 名称无效")
	require.ErrorContains(t, validateAgentDirectives([]xyRequest.AgentDirective{{Type: "tool", Name: "task_get"}, {Type: "tool", Name: "message_search"}}), "只能指定一次")
}

// TestLoadAgentHistoryIncludesRecentToolTrace 验证对应测试场景。
func TestLoadAgentHistoryIncludesRecentToolTrace(t *testing.T) {
	db := agentRunnerTestDB(t)
	session := model.AgentSession{OwnerID: 11, Title: "轨迹会话"}
	require.NoError(t, db.Create(&session).Error)
	run := model.AgentRun{OwnerID: 11, SessionID: session.ID, ModelID: 1, ModelName: "test", Status: "completed", ConfigSnapshot: "{}", StartedAt: time.Now()}
	require.NoError(t, db.Create(&run).Error)
	require.NoError(t, db.Create(&model.AgentRunTool{
		RunID: run.ID, CallID: "trace-1", ToolName: "message_search", Arguments: `{"dialog_id":7,"query":"Go"}`,
		Result: `{"status":"success","data":{"dialog":{"name":"Go 资料"},"messages":[]}}`, Status: "completed", StartedAt: time.Now(),
	}).Error)
	require.NoError(t, db.Create(&model.AgentMessage{OwnerID: 11, SessionID: session.ID, Sequence: 1, Role: "user", Content: "刚才搜索了哪里"}).Error)

	history, err := loadAgentHistory(t.Context(), db, 11, session.ID, 40)
	require.NoError(t, err)
	require.Len(t, history, 2)
	require.Equal(t, schema.System, history[0].Role)
	require.Contains(t, history[0].Content, "Go 资料")
	require.Equal(t, schema.User, history[1].Role)
}

// agentRunnerTestDB 提供测试辅助能力。
func agentRunnerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(
		&model.AIModel{}, &model.MCPServer{}, &model.AgentConfig{}, &model.AgentSession{},
		&model.AgentMessage{}, &model.AgentRun{}, &model.AgentRunTool{}, &model.AgentEvent{},
	))
	aiModel := model.AIModel{
		Name: "测试模型", Provider: "openai", BaseURL: "https://example.com/v1", APIKey: "secret-key",
		Model: "test", Enabled: true, Priority: 1, SupportsTools: true, SupportsStreaming: true,
		ContextWindow: 128000, MaxOutputTokens: 2000, RequestTimeoutSeconds: 60,
	}
	require.NoError(t, db.Create(&aiModel).Error)
	config := model.DefaultAgentConfig()
	require.NoError(t, db.Create(&config).Error)
	return db
}

// stubAgentModel 是测试辅助类型。
type stubAgentModel struct {
	answer string       // answer 是模拟模型回复。
	err    error        // err 是模拟模型错误。
	calls  atomic.Int64 // calls 记录模型调用次数。
}

// cancelAwareAgentModel 是测试辅助类型。
type cancelAwareAgentModel struct{}

// Generate 提供测试辅助能力。
func (cancelAwareAgentModel) Generate(context.Context, []*schema.Message, ...einomodel.Option) (*schema.Message, error) {
	return nil, errors.New("unexpected non-streaming call")
}

// Stream 提供测试辅助能力。
func (cancelAwareAgentModel) Stream(ctx context.Context, _ []*schema.Message, _ ...einomodel.Option) (*schema.StreamReader[*schema.Message], error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

// Generate 提供测试辅助能力。
func (m *stubAgentModel) Generate(_ context.Context, _ []*schema.Message, _ ...einomodel.Option) (*schema.Message, error) {
	return nil, errors.New("unexpected non-streaming call")
}

// Stream 提供测试辅助能力。
func (m *stubAgentModel) Stream(context.Context, []*schema.Message, ...einomodel.Option) (*schema.StreamReader[*schema.Message], error) {
	m.calls.Add(1)
	if m.err != nil {
		return nil, m.err
	}
	return schema.StreamReaderFromArray([]*schema.Message{schema.AssistantMessage(m.answer, nil)}), nil
}
