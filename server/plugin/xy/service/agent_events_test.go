package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAgentEventHubReplaysAfterLastEventID(t *testing.T) {
	hub := newAgentEventHub()
	first, err := hub.Publish(11, "run.started", map[string]any{"runId": 1})
	require.NoError(t, err)
	second, err := hub.Publish(11, "assistant.delta", map[string]any{"content": "你好"})
	require.NoError(t, err)
	require.EqualValues(t, 1, first.ID)
	require.EqualValues(t, 2, second.ID)

	replay, stream, cancel := hub.Subscribe(11, first.ID)
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
	_, stream, cancel := hub.Subscribe(22, 0)
	defer cancel()
	_, err := hub.Publish(11, "run.started", nil)
	require.NoError(t, err)
	select {
	case <-stream:
		t.Fatal("received an event from another session")
	default:
	}
	_, err = hub.Publish(11, "assistant.delta", string(make([]byte, agentEventMaxBytes)))
	require.ErrorContains(t, err, "内容过大")
}
