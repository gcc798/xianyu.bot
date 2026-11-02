package service

import (
	"errors"
	"testing"
	"time"

	"github.com/gcc798/go.xianyu.assistant/server/internal/infra"
	"github.com/gcc798/go.xianyu.assistant/server/internal/model/xianyu"
	xyRequest "github.com/gcc798/go.xianyu.assistant/server/internal/model/xianyu/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAgentSessionScopesConversationByOwner(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentSession{}, &model.AgentMessage{}, &model.AgentRun{}, &model.AgentRunTool{}, &model.AgentEvent{}, &model.ResourceTask{}))
	oldDB := infra.DB
	infra.DB = db
	t.Cleanup(func() { infra.DB = oldDB })

	sessions := new(agentSession)
	owned, err := sessions.Create(t.Context(), 11, xyRequest.CreateAgentSession{})
	require.NoError(t, err)
	require.Equal(t, "新对话", owned.Title)
	other, err := sessions.Create(t.Context(), 22, xyRequest.CreateAgentSession{Title: "私有会话"})
	require.NoError(t, err)

	now := time.Now()
	require.NoError(t, db.Model(&owned).Update("last_message_at", now).Error)
	require.NoError(t, db.Create(&[]model.AgentMessage{
		{OwnerID: 11, SessionID: owned.ID, Sequence: 2, Role: "assistant", Content: "第二条"},
		{OwnerID: 11, SessionID: owned.ID, Sequence: 1, Role: "user", Content: "第一条"},
		{OwnerID: 22, SessionID: other.ID, Sequence: 1, Role: "user", Content: "不可见"},
	}).Error)
	run := model.AgentRun{OwnerID: 11, SessionID: owned.ID, ModelID: 1, ModelName: "test", Status: "completed", ConfigSnapshot: `{}`, StartedAt: now}
	require.NoError(t, db.Create(&run).Error)
	require.NoError(t, db.Create(&model.AgentRunTool{RunID: run.ID, CallID: "approval-1", ToolName: "resource_acquire", Status: "waiting_approval", StartedAt: now}).Error)
	require.NoError(t, db.Create(&model.AgentRunTool{RunID: run.ID, CallID: "trace-1", ToolName: "message_search", Arguments: `{"query":"Go"}`, Result: `{"status":"success"}`, Status: "completed", StartedAt: now}).Error)
	require.NoError(t, db.Create(&[]model.AgentEvent{
		{OwnerID: 11, SessionID: owned.ID, RunID: new(run.ID), TurnNo: new(1), EventType: "turn.started", Payload: `{"runId":1,"turnNo":1}`, CreatedAt: now},
		{OwnerID: 11, SessionID: owned.ID, RunID: new(run.ID), TurnNo: new(1), StepNo: new(1), EventType: "tool.completed", CallID: "trace-1", Payload: `{"runId":1,"callId":"trace-1"}`, CreatedAt: now},
		{OwnerID: 22, SessionID: other.ID, EventType: "system.prompt", Payload: `{"content":"private"}`, CreatedAt: now},
	}).Error)
	require.NoError(t, db.Create(&model.ResourceTask{OwnerID: 11, AgentRunID: new(run.ID), CandidateID: 1, MCPServerCode: "resources", IdempotencyKey: "task-1", State: "running"}).Error)

	list, total, err := sessions.List(t.Context(), 11, xyRequest.AgentSessionSearch{})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, owned.ID, list[0].ID)

	detail, err := sessions.Get(t.Context(), 11, owned.ID)
	require.NoError(t, err)
	require.Len(t, detail.Messages, 2)
	require.Len(t, detail.PendingApprovals, 1)
	require.Equal(t, "resource_acquire", detail.PendingApprovals[0].ToolName)
	require.Len(t, detail.Tasks, 1)
	require.Len(t, detail.ToolTraces, 2)
	require.Equal(t, "message_search", detail.ToolTraces[1].ToolName)
	require.Equal(t, "trace-1", detail.ToolTraces[1].CallID)
	require.JSONEq(t, `{"status":"success"}`, detail.ToolTraces[1].Result)
	require.Equal(t, run.ID, detail.Tasks[0].RunID)
	require.EqualValues(t, 1, detail.Messages[0].Sequence)
	firstPage, err := sessions.Trajectory(t.Context(), 11, owned.ID, xyRequest.AgentTrajectorySearch{Limit: 1})
	require.NoError(t, err)
	require.Len(t, firstPage.Events, 1)
	require.True(t, firstPage.HasMore)
	require.JSONEq(t, `{"runId":1,"turnNo":1}`, string(firstPage.Events[0].Data))
	secondPage, err := sessions.Trajectory(t.Context(), 11, owned.ID, xyRequest.AgentTrajectorySearch{AfterID: firstPage.NextAfterID, Limit: 10})
	require.NoError(t, err)
	require.Len(t, secondPage.Events, 1)
	require.False(t, secondPage.HasMore)
	require.Equal(t, "trace-1", secondPage.Events[0].CallID)
	toolCall, err := sessions.ToolCall(t.Context(), 11, owned.ID, "trace-1")
	require.NoError(t, err)
	require.JSONEq(t, `{"query":"Go"}`, string(toolCall.Arguments))
	require.JSONEq(t, `{"status":"success"}`, string(toolCall.Result))
	_, err = sessions.ToolCall(t.Context(), 22, owned.ID, "trace-1")
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	_, err = sessions.Get(t.Context(), 11, other.ID)
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	_, err = sessions.Trajectory(t.Context(), 11, other.ID, xyRequest.AgentTrajectorySearch{})
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestAgentSessionDeleteRejectsRunningAndRemovesConversation(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentSession{}, &model.AgentMessage{}, &model.AgentRun{}, &model.AgentRunTool{}, &model.AgentEvent{}, &model.ResourceTask{}))
	oldDB := infra.DB
	infra.DB = db
	t.Cleanup(func() { infra.DB = oldDB })

	session := model.AgentSession{OwnerID: 11, Title: "待删除"}
	require.NoError(t, db.Create(&session).Error)
	run := model.AgentRun{OwnerID: 11, SessionID: session.ID, ModelID: 1, ModelName: "test", Status: "running", StartedAt: time.Now()}
	require.NoError(t, db.Create(&run).Error)
	require.NoError(t, db.Create(&model.AgentMessage{OwnerID: 11, SessionID: session.ID, RunID: &run.ID, Sequence: 1, Role: "user", Content: "内容"}).Error)
	require.NoError(t, db.Create(&model.AgentRunTool{RunID: run.ID, CallID: "call-1", ToolName: "resource_search", Status: "completed", StartedAt: time.Now()}).Error)

	sessions := new(agentSession)
	require.ErrorContains(t, sessions.Delete(t.Context(), 11, session.ID), "正在运行")
	require.NoError(t, db.Model(&run).Update("status", "completed").Error)
	require.NoError(t, sessions.Delete(t.Context(), 11, session.ID))

	for _, table := range []any{&model.AgentSession{}, &model.AgentMessage{}, &model.AgentRun{}, &model.AgentRunTool{}, &model.AgentEvent{}} {
		var count int64
		require.NoError(t, db.Model(table).Count(&count).Error)
		require.Zero(t, count)
	}
}
