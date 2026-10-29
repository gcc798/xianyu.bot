package runtime

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

var (
	ErrAlreadyRunning     = errors.New("账号通信运行时已启动")
	ErrNotConnected       = errors.New("账号通信运行时未连接")
	ErrQueueFull          = errors.New("账号消息队列已满")
	ErrReadTimeout        = errors.New("账号连接读取超时")
	ErrHeartbeatTimeout   = errors.New("账号连接心跳超时")
	ErrRequestUnsupported = errors.New("账号通信连接不支持请求响应")
)

type RequestConnection interface {
	Request(context.Context, []byte) ([]byte, error)
}

type State string

const (
	StateConnecting   State = "connecting"
	StateConnected    State = "connected"
	StateReconnecting State = "reconnecting"
	StateFailed       State = "failed"
)

type NonRetryableError interface {
	error
	NonRetryable() bool
}

type ActionError interface {
	error
	ActionURL() string
}

// Connection 的实现必须遵守 Context 取消，Close 必须能解除正在等待的 Read/Ping。
type Connection interface {
	Read(context.Context) ([]byte, error)
	Send(context.Context, []byte) error
	Ping(context.Context) error
	Close() error
}

type Dialer interface {
	Dial(context.Context, uint) (Connection, error)
}

type Handler func(context.Context, uint, []byte) error

type Config struct {
	QueueSize         int
	HeartbeatInterval time.Duration
	OperationTimeout  time.Duration
	ReadTimeout       time.Duration
	BackoffMin        time.Duration
	BackoffMax        time.Duration
	BackoffResetAfter time.Duration
	JitterPercent     uint
	StateChanged      func(Snapshot)
}

func DefaultConfig() Config {
	return Config{
		QueueSize:         64,
		HeartbeatInterval: 20 * time.Second,
		OperationTimeout:  10 * time.Second,
		ReadTimeout:       90 * time.Second,
		BackoffMin:        time.Second,
		BackoffMax:        30 * time.Second,
		BackoffResetAfter: time.Minute,
		JitterPercent:     20,
	}
}

func (c Config) validate() error {
	if c.QueueSize <= 0 || c.HeartbeatInterval <= 0 || c.OperationTimeout <= 0 || c.ReadTimeout <= 0 {
		return errors.New("队列容量和通信超时必须大于零")
	}
	if c.BackoffMin <= 0 || c.BackoffMax < c.BackoffMin || c.BackoffResetAfter <= 0 {
		return errors.New("重连退避配置无效")
	}
	if c.JitterPercent > 100 {
		return errors.New("重连抖动比例不能超过 100")
	}
	return nil
}

type Snapshot struct {
	AccountID  uint
	State      State
	QueueDepth int
	Reconnects uint64
	LastError  string
	ActionURL  string
}

type Manager struct {
	dialer  Dialer
	handler Handler
	config  Config

	mu        sync.Mutex
	runtimes  map[uint]*accountRuntime
	snapshots map[uint]Snapshot
}

func NewManager(dialer Dialer, handler Handler, config Config) (*Manager, error) {
	if dialer == nil || handler == nil {
		return nil, errors.New("通信 Dialer 和 Handler 不能为空")
	}
	if err := config.validate(); err != nil {
		return nil, err
	}
	return &Manager{
		dialer: dialer, handler: handler, config: config,
		runtimes: make(map[uint]*accountRuntime), snapshots: make(map[uint]Snapshot),
	}, nil
}

func (m *Manager) Start(ctx context.Context, accountID uint) error {
	if accountID == 0 {
		return errors.New("账号 ID 不能为空")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.runtimes[accountID]; exists {
		return ErrAlreadyRunning
	}
	delete(m.snapshots, accountID)
	runCtx, cancel := context.WithCancelCause(ctx)
	runtime := &accountRuntime{
		accountID: accountID,
		dialer:    m.dialer,
		handler:   m.handler,
		config:    m.config,
		cancel:    cancel,
		done:      make(chan struct{}),
		state:     StateConnecting,
	}
	m.runtimes[accountID] = runtime
	go func() {
		runtime.run(runCtx)
		m.mu.Lock()
		if m.runtimes[accountID] == runtime {
			delete(m.runtimes, accountID)
			m.snapshots[accountID] = runtime.snapshot()
		}
		m.mu.Unlock()
		close(runtime.done)
	}()
	return nil
}

func (m *Manager) Stop(accountID uint) bool {
	m.mu.Lock()
	runtime, exists := m.runtimes[accountID]
	m.mu.Unlock()
	if !exists {
		return false
	}
	runtime.cancel(context.Canceled)
	<-runtime.done
	m.mu.Lock()
	delete(m.snapshots, accountID)
	m.mu.Unlock()
	return true
}

func (m *Manager) Snapshot(accountID uint) (Snapshot, bool) {
	m.mu.Lock()
	runtime, exists := m.runtimes[accountID]
	snapshot := m.snapshots[accountID]
	m.mu.Unlock()
	if !exists {
		return snapshot, false
	}
	return runtime.snapshot(), true
}

func (m *Manager) Send(ctx context.Context, accountID uint, payload []byte) error {
	m.mu.Lock()
	runtime, exists := m.runtimes[accountID]
	m.mu.Unlock()
	if !exists {
		return ErrNotConnected
	}
	return runtime.send(ctx, payload)
}

func (m *Manager) Request(ctx context.Context, accountID uint, payload []byte) ([]byte, error) {
	m.mu.Lock()
	runtime, exists := m.runtimes[accountID]
	m.mu.Unlock()
	if !exists {
		return nil, ErrNotConnected
	}
	return runtime.request(ctx, payload)
}

type accountRuntime struct {
	accountID uint
	dialer    Dialer
	handler   Handler
	config    Config
	cancel    context.CancelCauseFunc
	done      chan struct{}

	mu         sync.Mutex
	state      State
	queueDepth int
	reconnects uint64
	lastError  string
	actionURL  string
	connection Connection
}

func (r *accountRuntime) run(ctx context.Context) {
	attempt := 0
	for {
		if context.Cause(ctx) != nil {
			return
		}
		r.setState(StateConnecting)
		connection, err := r.dialer.Dial(ctx, r.accountID)
		if err == nil {
			connectedAt := time.Now()
			r.setConnected()
			err = r.serve(ctx, connection)
			if time.Since(connectedAt) >= r.config.BackoffResetAfter {
				attempt = 0
			}
		}
		if context.Cause(ctx) != nil {
			return
		}
		if permanent, ok := errors.AsType[NonRetryableError](err); ok && permanent.NonRetryable() {
			r.setFailure(StateFailed, err, false)
			return
		}
		r.setFailure(StateReconnecting, err, true)
		if !wait(ctx, backoffDelay(attempt, r.config)) {
			return
		}
		attempt++
	}
}

func (r *accountRuntime) serve(ctx context.Context, connection Connection) error {
	r.mu.Lock()
	r.connection = connection
	r.mu.Unlock()
	defer func() {
		r.mu.Lock()
		if r.connection == connection {
			r.connection = nil
		}
		r.mu.Unlock()
	}()
	sessionCtx, cancel := context.WithCancel(ctx)
	messages := make(chan []byte, r.config.QueueSize)
	errorsCh := make(chan error, 3)
	var workers sync.WaitGroup
	fail := func(err error) {
		select {
		case errorsCh <- err:
		default:
		}
	}

	workers.Add(1)
	go func() {
		defer workers.Done()
		r.readLoop(sessionCtx, connection, messages, fail)
	}()
	workers.Add(1)
	go func() {
		defer workers.Done()
		r.handleLoop(sessionCtx, messages, fail)
	}()
	workers.Add(1)
	go func() {
		defer workers.Done()
		r.heartbeatLoop(sessionCtx, connection, fail)
	}()

	var err error
	select {
	case <-ctx.Done():
		err = context.Cause(ctx)
	case err = <-errorsCh:
	}
	cancel()
	_ = connection.Close()
	workers.Wait()
	r.setQueueDepth(0)
	return err
}

func (r *accountRuntime) send(ctx context.Context, payload []byte) error {
	r.mu.Lock()
	connection := r.connection
	r.mu.Unlock()
	if connection == nil {
		return ErrNotConnected
	}
	return connection.Send(ctx, bytes.Clone(payload))
}

func (r *accountRuntime) request(ctx context.Context, payload []byte) ([]byte, error) {
	r.mu.Lock()
	connection := r.connection
	r.mu.Unlock()
	if connection == nil {
		return nil, ErrNotConnected
	}
	requester, ok := connection.(RequestConnection)
	if !ok {
		return nil, ErrRequestUnsupported
	}
	return requester.Request(ctx, bytes.Clone(payload))
}

func (r *accountRuntime) readLoop(ctx context.Context, connection Connection, messages chan<- []byte, fail func(error)) {
	for {
		readCtx, cancel := context.WithTimeoutCause(ctx, r.config.ReadTimeout, ErrReadTimeout)
		message, err := connection.Read(readCtx)
		cause := context.Cause(readCtx)
		cancel()
		if err != nil {
			if context.Cause(ctx) != nil {
				return
			}
			if errors.Is(cause, ErrReadTimeout) {
				fail(ErrReadTimeout)
			} else {
				fail(fmt.Errorf("读取账号消息: %w", err))
			}
			return
		}
		if len(message) == 0 {
			continue
		}
		select {
		case messages <- bytes.Clone(message):
			r.setQueueDepth(len(messages))
		case <-ctx.Done():
			return
		default:
			fail(ErrQueueFull)
			return
		}
	}
}

func (r *accountRuntime) handleLoop(ctx context.Context, messages <-chan []byte, fail func(error)) {
	for {
		select {
		case <-ctx.Done():
			return
		case message := <-messages:
			r.setQueueDepth(len(messages))
			if err := r.handler(ctx, r.accountID, message); err != nil {
				fail(fmt.Errorf("处理账号消息: %w", err))
				return
			}
		}
	}
}

func (r *accountRuntime) heartbeatLoop(ctx context.Context, connection Connection, fail func(error)) {
	ticker := time.NewTicker(r.config.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeoutCause(ctx, r.config.OperationTimeout, ErrHeartbeatTimeout)
			err := connection.Ping(pingCtx)
			cause := context.Cause(pingCtx)
			cancel()
			if err == nil {
				continue
			}
			if context.Cause(ctx) != nil {
				return
			}
			if errors.Is(cause, ErrHeartbeatTimeout) {
				fail(ErrHeartbeatTimeout)
			} else {
				fail(fmt.Errorf("发送账号心跳: %w", err))
			}
			return
		}
	}
}

func (r *accountRuntime) snapshot() Snapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	return Snapshot{
		AccountID: r.accountID, State: r.state, QueueDepth: r.queueDepth,
		Reconnects: r.reconnects, LastError: r.lastError, ActionURL: r.actionURL,
	}
}

func (r *accountRuntime) setState(state State) {
	r.mu.Lock()
	r.state = state
	r.mu.Unlock()
	r.notify()
}

func (r *accountRuntime) setConnected() {
	r.mu.Lock()
	r.state = StateConnected
	r.lastError = ""
	r.actionURL = ""
	r.mu.Unlock()
	r.notify()
}

func (r *accountRuntime) setFailure(state State, err error, reconnect bool) {
	r.mu.Lock()
	r.state = state
	if reconnect {
		r.reconnects++
	}
	r.lastError = err.Error()
	r.actionURL = ""
	if action, ok := errors.AsType[ActionError](err); ok {
		r.actionURL = action.ActionURL()
	}
	r.mu.Unlock()
	r.notify()
}

func (r *accountRuntime) notify() {
	if r.config.StateChanged != nil {
		r.config.StateChanged(r.snapshot())
	}
}

func (r *accountRuntime) setQueueDepth(depth int) {
	r.mu.Lock()
	r.queueDepth = depth
	r.mu.Unlock()
}

func wait(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func backoffDelay(attempt int, config Config) time.Duration {
	delay := config.BackoffMin
	for range min(attempt, 62) {
		if delay >= config.BackoffMax/2 {
			delay = config.BackoffMax
			break
		}
		delay *= 2
	}
	delay = min(delay, config.BackoffMax)
	percent := time.Duration(config.JitterPercent)
	jitterMax := delay/100*percent + delay%100*percent/100
	if jitterMax == 0 {
		return delay
	}
	return delay - time.Duration(rand.Uint64()%uint64(jitterMax+1))
}
