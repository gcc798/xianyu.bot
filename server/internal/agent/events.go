package agent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	"gorm.io/gorm"
)

// agentEventMaxBytes 限制单个 Agent 事件的最大字节数。
const agentEventMaxBytes = 1 << 20

// Event 表示推送给会话订阅者的 Agent 事件。
type Event struct {
	ID        uint64          `json:"id"`        // ID 是事件序号。
	Type      string          `json:"type"`      // Type 是事件类型。
	Data      json.RawMessage `json:"data"`      // Data 是事件载荷。
	CreatedAt time.Time       `json:"createdAt"` // CreatedAt 是事件创建时间。
}

// agentEventSession 保存单个会话的内存事件状态。
type agentEventSession struct {
	nextID      uint64                  // nextID 是下一个内存事件序号。
	history     []Event                 // history 保存可重放的内存事件。
	subscribers map[chan Event]struct{} // subscribers 保存实时订阅者。
}

// EventHub 持久化稳定事件并广播持久化或瞬时事件。
// 数据库为空时仅在内存中保存事件，供隔离测试使用。
type EventHub struct {
	mu       sync.Mutex                  // mu 保护会话和订阅状态。
	db       *gorm.DB                    // db 持久化稳定事件。
	sessions map[uint]*agentEventSession // sessions 保存会话事件状态。
}

// NewEventHub 创建事件中心，可选数据库用于持久化。
func NewEventHub(databases ...*gorm.DB) *EventHub {
	var db *gorm.DB
	if len(databases) > 0 {
		db = databases[0]
	}
	return &EventHub{db: db, sessions: make(map[uint]*agentEventSession)}
}

// Publish 持久化并广播具有审计价值的事件。
func (h *EventHub) Publish(sessionID uint, eventType string, data any) (Event, error) {
	payload, err := marshalAgentEvent(sessionID, eventType, data)
	if err != nil {
		return Event{}, err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.database() == nil {
		return h.publishInMemory(sessionID, eventType, payload), nil
	}

	stored, err := h.persist(sessionID, eventType, payload)
	if err != nil {
		return Event{}, err
	}
	event := Event{ID: stored.ID, Type: stored.EventType, Data: json.RawMessage(stored.Payload), CreatedAt: stored.CreatedAt}
	h.broadcast(sessionID, event)
	return event, nil
}

// PublishTransient 广播无需独立审计的瞬时事件，如流式文本片段。
func (h *EventHub) PublishTransient(sessionID uint, eventType string, data any) (Event, error) {
	payload, err := marshalAgentEvent(sessionID, eventType, data)
	if err != nil {
		return Event{}, err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	event := Event{Type: eventType, Data: payload, CreatedAt: time.Now()}
	for subscriber := range h.session(sessionID).subscribers {
		if !sendAgentEvent(subscriber, event) {
			delete(h.session(sessionID).subscribers, subscriber)
		}
	}
	return event, nil
}

// Subscribe 订阅会话事件，并返回指定序号后的历史事件。
func (h *EventHub) Subscribe(ctx context.Context, sessionID uint, afterID uint64) ([]Event, <-chan Event, func(), error) {
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
	stream := make(chan Event, 64)
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

// Forget 关闭会话订阅并清除内存状态。
func (h *EventHub) Forget(sessionID uint) {
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

// persist 将事件及其运行引用写入数据库。
func (h *EventHub) persist(sessionID uint, eventType string, payload []byte) (model.AgentEvent, error) {
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

// replay 读取指定序号后的持久化或内存事件。
func (h *EventHub) replay(ctx context.Context, sessionID uint, afterID uint64) ([]Event, error) {
	if db := h.database(); db != nil {
		stored := make([]model.AgentEvent, 0)
		if err := db.WithContext(ctx).Where("session_id = ? AND id > ?", sessionID, afterID).Order("id").Find(&stored).Error; err != nil {
			return nil, err
		}
		replay := make([]Event, 0, len(stored))
		for _, event := range stored {
			replay = append(replay, Event{ID: event.ID, Type: event.EventType, Data: json.RawMessage(event.Payload), CreatedAt: event.CreatedAt})
		}
		return replay, nil
	}

	replay := make([]Event, 0)
	for _, event := range h.session(sessionID).history {
		if event.ID > afterID {
			replay = append(replay, event)
		}
	}
	return replay, nil
}

// publishInMemory 保存并广播内存事件。
func (h *EventHub) publishInMemory(sessionID uint, eventType string, payload []byte) Event {
	session := h.session(sessionID)
	session.nextID++
	event := Event{ID: session.nextID, Type: eventType, Data: payload, CreatedAt: time.Now()}
	session.history = append(session.history, event)
	for subscriber := range session.subscribers {
		if !sendAgentEvent(subscriber, event) {
			delete(session.subscribers, subscriber)
		}
	}
	return event
}

// broadcast 向会话的全部有效订阅者广播事件。
func (h *EventHub) broadcast(sessionID uint, event Event) {
	session := h.session(sessionID)
	for subscriber := range session.subscribers {
		if !sendAgentEvent(subscriber, event) {
			delete(session.subscribers, subscriber)
		}
	}
}

// database 返回事件持久化数据库。
func (h *EventHub) database() *gorm.DB {
	return h.db
}

// session 获取或创建会话事件状态。
func (h *EventHub) session(sessionID uint) *agentEventSession {
	session := h.sessions[sessionID]
	if session == nil {
		session = &agentEventSession{subscribers: make(map[chan Event]struct{})}
		h.sessions[sessionID] = session
	}
	return session
}

// marshalAgentEvent 校验事件标识并序列化载荷。
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

// agentEventReferences 保存事件关联的运行上下文。
type agentEventReferences struct {
	RunID        *uint  // RunID 是运行标识。
	TurnNo       *int   // TurnNo 是轮次编号。
	StepNo       *int   // StepNo 是步骤编号。
	CallID       string // CallID 是工具调用标识。
	ParentCallID string // ParentCallID 是父调用标识。
}

// eventReferences 从事件载荷提取运行上下文。
func eventReferences(payload []byte) agentEventReferences {
	var raw struct {
		RunID        uint   `json:"runId"`        // RunID 是运行标识。
		TurnNo       int    `json:"turnNo"`       // TurnNo 是轮次编号。
		StepNo       int    `json:"stepNo"`       // StepNo 是步骤编号。
		CallID       string `json:"callId"`       // CallID 是工具调用标识。
		ParentCallID string `json:"parentCallId"` // ParentCallID 是父调用标识。
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

// sendAgentEvent 非阻塞发送事件，订阅者拥塞时将其关闭。
func sendAgentEvent(subscriber chan Event, event Event) bool {
	select {
	case subscriber <- event:
		return true
	default:
		close(subscriber)
		return false
	}
}
