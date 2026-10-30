package service

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model"
	xyRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAgentRunnerPersistsConversationAndDeduplicatesRequest(t *testing.T) {
	db := agentRunnerTestDB(t)
	session := model.AgentSession{OwnerID: 11, Title: "测试会话"}
	require.NoError(t, db.Create(&session).Error)
	chatModel := &stubAgentModel{answer: "这是最终回答"}
	runner := &agentRunner{newModel: func(context.Context, model.AIModel, model.AgentConfig) (einomodel.BaseChatModel, error) {
		return chatModel, nil
	}}

	result, err := runner.Run(t.Context(), 11, session.ID, "帮我找 Go 资料", "request-1")
	require.NoError(t, err)
	require.Equal(t, "completed", result.Run.Status)
	require.NotNil(t, result.Assistant)
	require.Equal(t, "这是最终回答", result.Assistant.Content)
	require.EqualValues(t, 1, chatModel.calls.Load())
	require.NotContains(t, result.Run.ConfigSnapshot, "secret-key")

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

func TestAgentRunnerRecordsModelFailure(t *testing.T) {
	db := agentRunnerTestDB(t)
	session := model.AgentSession{OwnerID: 11, Title: "失败会话"}
	require.NoError(t, db.Create(&session).Error)
	runner := &agentRunner{newModel: func(context.Context, model.AIModel, model.AgentConfig) (einomodel.BaseChatModel, error) {
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
}

func TestAgentRunnerSubmitPublishesStreamingEvents(t *testing.T) {
	db := agentRunnerTestDB(t)
	session := model.AgentSession{OwnerID: 11, Title: "异步会话"}
	require.NoError(t, db.Create(&session).Error)
	events := newAgentEventHub()
	runner := &agentRunner{events: events, newModel: func(context.Context, model.AIModel, model.AgentConfig) (einomodel.BaseChatModel, error) {
		return &stubAgentModel{answer: "流式回答"}, nil
	}}
	_, stream, cancel := events.Subscribe(session.ID, 0)
	defer cancel()

	accepted, err := runner.Submit(t.Context(), 11, session.ID, "异步问题", "request-async")
	require.NoError(t, err)
	require.Equal(t, "running", accepted.Run.Status)

	types := make([]string, 0, 3)
	deadline := time.After(3 * time.Second)
	for len(types) < 3 {
		select {
		case event := <-stream:
			types = append(types, event.Type)
		case <-deadline:
			t.Fatalf("timed out waiting for events: %v", types)
		}
	}
	require.Equal(t, []string{"run.started", "assistant.delta", "run.completed"}, types)
	var stored model.AgentRun
	require.NoError(t, db.First(&stored, accepted.Run.ID).Error)
	require.Equal(t, "completed", stored.Status)
}

func TestAgentRunnerCancelsOwnedRun(t *testing.T) {
	db := agentRunnerTestDB(t)
	session := model.AgentSession{OwnerID: 11, Title: "取消会话"}
	require.NoError(t, db.Create(&session).Error)
	events := newAgentEventHub()
	runner := &agentRunner{events: events, newModel: func(context.Context, model.AIModel, model.AgentConfig) (einomodel.BaseChatModel, error) {
		return cancelAwareAgentModel{}, nil
	}}
	_, stream, unsubscribe := events.Subscribe(session.ID, 0)
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

func agentRunnerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.AutoMigrate(
		&model.AIModel{}, &model.MCPServer{}, &model.AgentConfig{}, &model.AgentSession{},
		&model.AgentMessage{}, &model.AgentRun{}, &model.AgentRunTool{},
	))
	oldDB := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = oldDB })
	aiModel := model.AIModel{
		Name: "测试模型", Provider: "openai", BaseURL: "https://example.com/v1", APIKey: "secret-key",
		Model: "test", Enabled: true, IsDefault: true, SupportsTools: true, SupportsStreaming: true,
		ContextWindow: 128000, MaxOutputTokens: 2000, RequestTimeoutSeconds: 60,
	}
	require.NoError(t, db.Create(&aiModel).Error)
	config := model.DefaultAgentConfig()
	require.NoError(t, db.Create(&config).Error)
	return db
}

type stubAgentModel struct {
	answer string
	err    error
	calls  atomic.Int64
}

type cancelAwareAgentModel struct{}

func (cancelAwareAgentModel) Generate(context.Context, []*schema.Message, ...einomodel.Option) (*schema.Message, error) {
	return nil, errors.New("unexpected non-streaming call")
}

func (cancelAwareAgentModel) Stream(ctx context.Context, _ []*schema.Message, _ ...einomodel.Option) (*schema.StreamReader[*schema.Message], error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func (m *stubAgentModel) Generate(_ context.Context, _ []*schema.Message, _ ...einomodel.Option) (*schema.Message, error) {
	return nil, errors.New("unexpected non-streaming call")
}

func (m *stubAgentModel) Stream(context.Context, []*schema.Message, ...einomodel.Option) (*schema.StreamReader[*schema.Message], error) {
	m.calls.Add(1)
	if m.err != nil {
		return nil, m.err
	}
	return schema.StreamReaderFromArray([]*schema.Message{schema.AssistantMessage(m.answer, nil)}), nil
}
