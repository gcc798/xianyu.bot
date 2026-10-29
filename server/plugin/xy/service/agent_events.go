package service

import (
	"encoding/json"
	"errors"
	"sync"
	"time"
)

const (
	agentEventHistoryLimit = 200
	agentEventMaxBytes     = 64 << 10
)

type agentEvent struct {
	ID        uint64          `json:"id"`
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	CreatedAt time.Time       `json:"createdAt"`
}

type agentEventSession struct {
	nextID      uint64
	history     []agentEvent
	subscribers map[chan agentEvent]struct{}
}

type agentEventHub struct {
	mu       sync.Mutex
	sessions map[uint]*agentEventSession
}

func newAgentEventHub() *agentEventHub {
	return &agentEventHub{sessions: make(map[uint]*agentEventSession)}
}

func (h *agentEventHub) Publish(sessionID uint, eventType string, data any) (agentEvent, error) {
	if sessionID == 0 || eventType == "" {
		return agentEvent{}, errors.New("Agent 事件缺少会话或类型")
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return agentEvent{}, err
	}
	if len(payload) > agentEventMaxBytes {
		return agentEvent{}, errors.New("Agent 事件内容过大")
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	session := h.session(sessionID)
	session.nextID++
	event := agentEvent{ID: session.nextID, Type: eventType, Data: payload, CreatedAt: time.Now()}
	session.history = append(session.history, event)
	if overflow := len(session.history) - agentEventHistoryLimit; overflow > 0 {
		session.history = append(session.history[:0], session.history[overflow:]...)
	}
	for subscriber := range session.subscribers {
		select {
		case subscriber <- event:
		default:
			close(subscriber)
			delete(session.subscribers, subscriber)
		}
	}
	return event, nil
}

func (h *agentEventHub) Subscribe(sessionID uint, afterID uint64) ([]agentEvent, <-chan agentEvent, func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	session := h.session(sessionID)
	replay := make([]agentEvent, 0)
	for _, event := range session.history {
		if event.ID > afterID {
			replay = append(replay, event)
		}
	}
	stream := make(chan agentEvent, 64)
	session.subscribers[stream] = struct{}{}
	cancel := sync.OnceFunc(func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if _, exists := session.subscribers[stream]; exists {
			delete(session.subscribers, stream)
			close(stream)
		}
	})
	return replay, stream, cancel
}

func (h *agentEventHub) session(sessionID uint) *agentEventSession {
	session := h.sessions[sessionID]
	if session == nil {
		session = &agentEventSession{subscribers: make(map[chan agentEvent]struct{})}
		h.sessions[sessionID] = session
	}
	return session
}
