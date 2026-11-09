package connection

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
	// ErrAlreadyRunning 表示账号运行时已存在。
	ErrAlreadyRunning = errors.New("账号通信运行时已启动")
	// ErrNotConnected 表示账号没有可用连接。
	ErrNotConnected = errors.New("账号通信运行时未连接")
	// ErrQueueFull 表示处理速度落后且消息队列已满。
	ErrQueueFull = errors.New("账号消息队列已满")
	// ErrReadTimeout 表示连接在读超时内没有收到协议帧。
	ErrReadTimeout = errors.New("账号连接读取超时")
	// ErrHeartbeatTimeout 表示心跳未在操作超时内完成。
	ErrHeartbeatTimeout = errors.New("账号连接心跳超时")
	// ErrRequestUnsupported 表示连接不支持请求响应关联。
	ErrRequestUnsupported = errors.New("账号通信连接不支持请求响应")
)

// RequestConnection 扩展支持请求响应模式的连接。
type RequestConnection interface {
	// Request 发送 payload 并等待关联响应。
	Request(context.Context, []byte) ([]byte, error)
}

// State 表示账号通信运行时状态。
type State string

const (
	// StateConnecting 表示正在建立首次连接。
	StateConnecting State = "connecting"
	// StateConnected 表示连接可用。
	StateConnected State = "connected"
	// StateReconnecting 表示连接失败后正在退避重连。
	StateReconnecting State = "reconnecting"
	// StateFailed 表示遇到不可重试错误并停止。
	StateFailed State = "failed"
)

// NonRetryableError 标记自动重连无法恢复的错误。
type NonRetryableError interface {
	error
	// NonRetryable 返回错误是否应终止运行时。
	NonRetryable() bool
}

// ActionError 为需要用户操作的错误提供安全跳转地址。
type ActionError interface {
	error
	// ActionURL 返回用户可访问的处理地址。
	ActionURL() string
}

// Connection 是账号通信的最小双向契约；Close 必须解除等待中的 Read 和 Ping。
type Connection interface {
	// Read 阻塞读取下一条业务消息，必须遵守上下文取消。
	Read(context.Context) ([]byte, error)
	// Send 发送一条业务消息，必须遵守上下文取消。
	Send(context.Context, []byte) error
	// Ping 执行一次连接级心跳。
	Ping(context.Context) error
	// Close 关闭连接并解除阻塞操作。
	Close() error
}

// Dialer 为指定账号建立已完成协议初始化的连接。
type Dialer interface {
	// Dial 建立 accountID 的新连接。
	Dial(context.Context, uint) (Connection, error)
}

// Handler 处理指定账号收到的一条完整业务消息。
type Handler func(context.Context, uint, []byte) error

// Config 定义连接队列、超时、重连和状态通知策略。
type Config struct {
	QueueSize         int            // QueueSize 是待读消息缓冲容量。
	HeartbeatInterval time.Duration  // HeartbeatInterval 是心跳间隔。
	OperationTimeout  time.Duration  // OperationTimeout 是单次发送或心跳超时。
	ReadTimeout       time.Duration  // ReadTimeout 是等待下一协议帧的最长时间。
	BackoffMin        time.Duration  // BackoffMin 是首次重连等待时间。
	BackoffMax        time.Duration  // BackoffMax 是重连等待上限。
	BackoffResetAfter time.Duration  // BackoffResetAfter 是连接稳定后重置退避的时长。
	JitterPercent     uint           // JitterPercent 是向下随机抖动百分比。
	StateChanged      func(Snapshot) // StateChanged 在运行时状态变化后同步调用。
}

// DefaultConfig 返回适合闲鱼长连接的保守默认参数。
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

// validate 校验队列、超时和退避参数边界。
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

// Snapshot 是账号运行时可安全对外读取的状态副本。
type Snapshot struct {
	AccountID  uint   // AccountID 是闲鱼账号 ID。
	State      State  // State 是当前连接状态。
	QueueDepth int    // QueueDepth 是待读消息数量。
	Reconnects uint64 // Reconnects 是累计重连次数。
	LastError  string // LastError 是最近一次连接错误。
	ActionURL  string // ActionURL 是错误要求的用户处理地址。
}

// Manager 管理多个账号彼此隔离的通信运行时。
type Manager struct {
	dialer  Dialer  // dialer 建立账号连接。
	handler Handler // handler 处理账号业务消息。
	config  Config  // config 是所有账号共享的运行策略。

	mu        sync.Mutex               // mu 保护 runtimes 和 snapshots。
	runtimes  map[uint]*accountRuntime // runtimes 保存活动账号运行时。
	snapshots map[uint]Snapshot        // snapshots 保存已终止运行时的最后状态。
}

// NewManager 校验依赖和配置后创建连接管理器。
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

// Start 为 accountID 启动唯一后台运行时。
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

// Stop 取消并等待账号运行时退出；未运行时返回 false。
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

// Snapshot 返回账号最新状态及其是否仍在运行。
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

// Send 复制 payload 后通过账号当前连接发送。
func (m *Manager) Send(ctx context.Context, accountID uint, payload []byte) error {
	m.mu.Lock()
	runtime, exists := m.runtimes[accountID]
	m.mu.Unlock()
	if !exists {
		return ErrNotConnected
	}
	return runtime.send(ctx, payload)
}

// Request 复制 payload 后通过支持请求响应的账号连接调用。
func (m *Manager) Request(ctx context.Context, accountID uint, payload []byte) ([]byte, error) {
	m.mu.Lock()
	runtime, exists := m.runtimes[accountID]
	m.mu.Unlock()
	if !exists {
		return nil, ErrNotConnected
	}
	return runtime.request(ctx, payload)
}

// accountRuntime 维护单个账号的连接、并发循环和重连状态。
type accountRuntime struct {
	accountID uint                    // accountID 是运行时所属账号 ID。
	dialer    Dialer                  // dialer 建立新连接。
	handler   Handler                 // handler 消费业务消息。
	config    Config                  // config 控制超时、队列和退避。
	cancel    context.CancelCauseFunc // cancel 停止运行时并记录原因。
	done      chan struct{}           // done 在运行时完全退出后关闭。

	mu         sync.Mutex // mu 保护下列状态与当前连接。
	state      State      // state 是当前连接状态。
	queueDepth int        // queueDepth 是待读消息数量。
	reconnects uint64     // reconnects 是累计重连次数。
	lastError  string     // lastError 是最近一次连接错误。
	actionURL  string     // actionURL 是错误要求的用户处理地址。
	connection Connection // connection 是当前可用连接。
}

// run 持续拨号和服务连接，遇到永久错误或上下文取消时退出。
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

// serve 并行运行读取、处理和心跳循环，任一失败后关闭全部循环。
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

// send 通过当前连接发送 payload 副本。
func (r *accountRuntime) send(ctx context.Context, payload []byte) error {
	r.mu.Lock()
	connection := r.connection
	r.mu.Unlock()
	if connection == nil {
		return ErrNotConnected
	}
	return connection.Send(ctx, bytes.Clone(payload))
}

// request 通过当前连接执行请求响应调用。
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

// readLoop 限时读取消息并在队列满时触发连接重建。
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

// handleLoop 顺序消费消息，处理失败时终止当前连接。
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

// heartbeatLoop 周期发送限时心跳，失败时触发连接重建。
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

// snapshot 在线程安全的临界区内复制运行时状态。
func (r *accountRuntime) snapshot() Snapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	return Snapshot{
		AccountID: r.accountID, State: r.state, QueueDepth: r.queueDepth,
		Reconnects: r.reconnects, LastError: r.lastError, ActionURL: r.actionURL,
	}
}

// setState 更新状态并通知观察者。
func (r *accountRuntime) setState(state State) {
	r.mu.Lock()
	r.state = state
	r.mu.Unlock()
	r.notify()
}

// setConnected 清除上次错误并标记连接可用。
func (r *accountRuntime) setConnected() {
	r.mu.Lock()
	r.state = StateConnected
	r.lastError = ""
	r.actionURL = ""
	r.mu.Unlock()
	r.notify()
}

// setFailure 保存错误、可选增加重连计数并提取用户操作地址。
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

// notify 向配置的观察者发送一致状态快照。
func (r *accountRuntime) notify() {
	if r.config.StateChanged != nil {
		r.config.StateChanged(r.snapshot())
	}
}

// setQueueDepth 更新待读消息数量。
func (r *accountRuntime) setQueueDepth(depth int) {
	r.mu.Lock()
	r.queueDepth = depth
	r.mu.Unlock()
}

// wait 等待 delay，ctx 取消时提前返回 false。
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

// backoffDelay 计算带向下随机抖动的封顶指数退避时长。
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
