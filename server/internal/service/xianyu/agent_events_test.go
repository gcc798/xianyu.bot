package service

import (
	"testing"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAgentEventHubReplaysAfterLastEventID(t *testing.T) {
	hub := newAgentEventHub()
	first, err := hub.Publish(11, "run.started", map[string]any{"runId": 1})
	require.NoError(t, err)
	second, err := hub.Publish(11, "assistant.delta", map[string]any{"content": "你好"})
	require.NoError(t, err)
	require.EqualValues(t, 1, first.ID)
	require.EqualValues(t, 2, second.ID)

	replay, stream, cancel, err := hub.Subscribe(t.Context(), 11, first.ID)
	require.NoError(t, err)
	defer cancel()
	require.Len(t, replay, 1)
	require.Equal(t, "assistant.delta", replay[0].Type)

	third, err := hub.Publish(11, "run.completed", map[string]any{"runId": 1})
	require.NoError(t, err)
	require.Equal(t, third, <-stream)
	hub.Forget(11)
	_, open := <-stream
	require.False(t, open)
}

func TestAgentEventHubSeparatesSessionsAndRejectsOversizedData(t *testing.T) {
	hub := newAgentEventHub()
	_, stream, cancel, err := hub.Subscribe(t.Context(), 22, 0)
	require.NoError(t, err)
	defer cancel()
	_, err = hub.Publish(11, "run.started", nil)
	require.NoError(t, err)
	select {
	case <-stream:
		t.Fatal("received an event from another session")
	default:
	}
	_, err = hub.Publish(11, "assistant.delta", string(make([]byte, agentEventMaxBytes)))
	require.ErrorContains(t, err, "内容过大")
}

func TestAgentEventHubPersistsAndReplaysAfterRestart(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentSession{}, &model.AgentEvent{}))
	session := model.AgentSession{OwnerID: 7, Title: "审计会话"}
	require.NoError(t, db.Create(&session).Error)

	firstHub := newAgentEventHub(db)
	stored, err := firstHub.Publish(session.ID, "run.started", map[string]any{"runId": 9, "turnNo": 1})
	require.NoError(t, err)
	require.NotZero(t, stored.ID)

	restartedHub := newAgentEventHub(db)
	replay, _, cancel, err := restartedHub.Subscribe(t.Context(), session.ID, 0)
	require.NoError(t, err)
	defer cancel()
	require.Len(t, replay, 1)
	require.Equal(t, stored.ID, replay[0].ID)
	require.Equal(t, "run.started", replay[0].Type)

	var event model.AgentEvent
	require.NoError(t, db.First(&event, stored.ID).Error)
	require.Equal(t, uint(7), event.OwnerID)
	require.Equal(t, new(uint(9)), event.RunID)
	require.Equal(t, new(1), event.TurnNo)
}

func TestAgentEventHubDoesNotPersistStreamingDelta(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentSession{}, &model.AgentEvent{}))
	session := model.AgentSession{OwnerID: 7, Title: "流式会话"}
	require.NoError(t, db.Create(&session).Error)

	hub := newAgentEventHub(db)
	_, stream, cancel, err := hub.Subscribe(t.Context(), session.ID, 0)
	require.NoError(t, err)
	defer cancel()
	event, err := hub.PublishTransient(session.ID, "assistant.delta", map[string]any{"content": "片段"})
	require.NoError(t, err)
	require.Zero(t, event.ID)
	require.Equal(t, event, <-stream)

	var count int64
	require.NoError(t, db.Model(&model.AgentEvent{}).Count(&count).Error)
	require.Zero(t, count)
}
