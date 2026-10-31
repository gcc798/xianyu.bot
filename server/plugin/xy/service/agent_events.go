package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model"
	"gorm.io/gorm"
)

const agentEventMaxBytes = 64 << 10

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

// agentEventHub persists settled Agent facts and broadcasts both persisted and transient events.
// The database is optional only for isolated unit tests; production resolves global.GVA_DB.
type agentEventHub struct {
	mu       sync.Mutex
	db       *gorm.DB
	sessions map[uint]*agentEventSession
}

func newAgentEventHub(databases ...*gorm.DB) *agentEventHub {
	var db *gorm.DB
	if len(databases) > 0 {
		db = databases[0]
	}
	return &agentEventHub{db: db, sessions: make(map[uint]*agentEventSession)}
}

func (h *agentEventHub) Publish(sessionID uint, eventType string, data any) (agentEvent, error) {
	payload, err := marshalAgentEvent(sessionID, eventType, data)
	if err != nil {
		return agentEvent{}, err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.database() == nil {
		return h.publishInMemory(sessionID, eventType, payload), nil
	}

	stored, err := h.persist(sessionID, eventType, payload)
	if err != nil {
		return agentEvent{}, err
	}
	event := agentEvent{ID: stored.ID, Type: stored.EventType, Data: json.RawMessage(stored.Payload), CreatedAt: stored.CreatedAt}
	h.broadcast(sessionID, event)
	return event, nil
}

// PublishTransient broadcasts data that has no standalone audit value, such as one streaming text delta.
func (h *agentEventHub) PublishTransient(sessionID uint, eventType string, data any) (agentEvent, error) {
	payload, err := marshalAgentEvent(sessionID, eventType, data)
	if err != nil {
		return agentEvent{}, err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	event := agentEvent{Type: eventType, Data: payload, CreatedAt: time.Now()}
	for subscriber := range h.session(sessionID).subscribers {
		if !sendAgentEvent(subscriber, event) {
			delete(h.session(sessionID).subscribers, subscriber)
		}
	}
	return event, nil
}

func (h *agentEventHub) Subscribe(ctx context.Context, sessionID uint, afterID uint64) ([]agentEvent, <-chan agentEvent, func(), error) {
	if sessionID == 0 {
		return nil, nil, nil, errors.New("Agent 事件缺少会话")
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	replay, err := h.replay(ctx, sessionID, afterID)
	if err != nil {
		return nil, nil, nil, err
	}
	session := h.session(sessionID)
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
	return replay, stream, cancel, nil
}

func (h *agentEventHub) Forget(sessionID uint) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if session := h.sessions[sessionID]; session != nil {
		for subscriber := range session.subscribers {
			close(subscriber)
			delete(session.subscribers, subscriber)
		}
		delete(h.sessions, sessionID)
	}
}

func (h *agentEventHub) persist(sessionID uint, eventType string, payload []byte) (model.AgentEvent, error) {
	db := h.database()
	var session model.AgentSession
	if err := db.Select("owner_id").First(&session, sessionID).Error; err != nil {
		return model.AgentEvent{}, err
	}
	references := eventReferences(payload)
	stored := model.AgentEvent{
		OwnerID: session.OwnerID, SessionID: sessionID, RunID: references.RunID,
		TurnNo: references.TurnNo, StepNo: references.StepNo, EventType: eventType,
		CallID: references.CallID, ParentCallID: references.ParentCallID,
		Payload: string(payload), CreatedAt: time.Now(),
	}
	if err := db.Create(&stored).Error; err != nil {
		return model.AgentEvent{}, err
	}
	return stored, nil
}

func (h *agentEventHub) replay(ctx context.Context, sessionID uint, afterID uint64) ([]agentEvent, error) {
	if db := h.database(); db != nil {
		stored := make([]model.AgentEvent, 0)
		if err := db.WithContext(ctx).Where("session_id = ? AND id > ?", sessionID, afterID).Order("id").Find(&stored).Error; err != nil {
			return nil, err
		}
		replay := make([]agentEvent, 0, len(stored))
		for _, event := range stored {
			replay = append(replay, agentEvent{ID: event.ID, Type: event.EventType, Data: json.RawMessage(event.Payload), CreatedAt: event.CreatedAt})
		}
		return replay, nil
	}

	replay := make([]agentEvent, 0)
	for _, event := range h.session(sessionID).history {
		if event.ID > afterID {
			replay = append(replay, event)
		}
	}
	return replay, nil
}

func (h *agentEventHub) publishInMemory(sessionID uint, eventType string, payload []byte) agentEvent {
	session := h.session(sessionID)
	session.nextID++
	event := agentEvent{ID: session.nextID, Type: eventType, Data: payload, CreatedAt: time.Now()}
	session.history = append(session.history, event)
	for subscriber := range session.subscribers {
		if !sendAgentEvent(subscriber, event) {
			delete(session.subscribers, subscriber)
		}
	}
	return event
}

func (h *agentEventHub) broadcast(sessionID uint, event agentEvent) {
	session := h.session(sessionID)
	for subscriber := range session.subscribers {
		if !sendAgentEvent(subscriber, event) {
			delete(session.subscribers, subscriber)
		}
	}
}

func (h *agentEventHub) database() *gorm.DB {
	if h.db != nil {
		return h.db
	}
	return global.GVA_DB
}

func (h *agentEventHub) session(sessionID uint) *agentEventSession {
	session := h.sessions[sessionID]
	if session == nil {
		session = &agentEventSession{subscribers: make(map[chan agentEvent]struct{})}
		h.sessions[sessionID] = session
	}
	return session
}

func marshalAgentEvent(sessionID uint, eventType string, data any) ([]byte, error) {
	if sessionID == 0 || eventType == "" || strings.ContainsAny(eventType, "\r\n") {
		return nil, errors.New("Agent 事件缺少会话或类型")
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if len(payload) > agentEventMaxBytes {
		return nil, errors.New("Agent 事件内容过大")
	}
	return payload, nil
}

type agentEventReferences struct {
	RunID        *uint
	TurnNo       *int
	StepNo       *int
	CallID       string
	ParentCallID string
}

func eventReferences(payload []byte) agentEventReferences {
	var raw struct {
		RunID        uint   `json:"runId"`
		TurnNo       int    `json:"turnNo"`
		StepNo       int    `json:"stepNo"`
		CallID       string `json:"callId"`
		ParentCallID string `json:"parentCallId"`
	}
	if json.Unmarshal(payload, &raw) != nil {
		return agentEventReferences{}
	}
	references := agentEventReferences{CallID: raw.CallID, ParentCallID: raw.ParentCallID}
	if raw.RunID > 0 {
		references.RunID = new(raw.RunID)
	}
	if raw.TurnNo > 0 {
		references.TurnNo = new(raw.TurnNo)
	}
	if raw.StepNo > 0 {
		references.StepNo = new(raw.StepNo)
	}
	return references
}

func sendAgentEvent(subscriber chan agentEvent, event agentEvent) bool {
	select {
	case subscriber <- event:
		return true
	default:
		close(subscriber)
		return false
	}
}
