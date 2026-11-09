package connection

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestManagerRunsOneCancelableRuntimePerAccount 验证对应测试场景。
func TestManagerRunsOneCancelableRuntimePerAccount(t *testing.T) {
	connection := newFakeConnection()
	dialer := &fakeDialer{connections: make(chan Connection, 1)}
	dialer.connections <- connection
	type receivedMessage struct {
		accountID uint
		body      []byte
	}
	received := make(chan receivedMessage, 1)
	manager, err := NewManager(dialer, func(_ context.Context, accountID uint, message []byte) error {
		received <- receivedMessage{accountID: accountID, body: message}
		return nil
	}, testConfig())
	require.NoError(t, err)
	require.NoError(t, manager.Start(t.Context(), 7))
	require.ErrorIs(t, manager.Start(t.Context(), 7), ErrAlreadyRunning)
	require.Eventually(t, func() bool {
		snapshot, ok := manager.Snapshot(7)
		return ok && snapshot.State == StateConnected
	}, time.Second, time.Millisecond)
	require.NoError(t, manager.Send(t.Context(), 7, []byte("outgoing")))
	require.Equal(t, []byte("outgoing"), <-connection.sends)

	message := []byte("message")
	connection.reads <- message
	got := <-received
	require.Equal(t, uint(7), got.accountID)
	require.Equal(t, message, got.body)
	message[0] = 'X'
	require.Equal(t, []byte("message"), got.body)
	require.Eventually(t, func() bool { return connection.pings.Load() > 0 }, time.Second, time.Millisecond)
	require.True(t, manager.Stop(7))
	require.False(t, manager.Stop(7))
	_, running := manager.Snapshot(7)
	require.False(t, running)
	require.True(t, connection.isClosed())
}

// TestRuntimeReconnectsWhenQueueIsFull 验证对应测试场景。
func TestRuntimeReconnectsWhenQueueIsFull(t *testing.T) {
	first := newFakeConnection()
	second := newFakeConnection()
	dialer := &fakeDialer{connections: make(chan Connection, 2)}
	dialer.connections <- first
	dialer.connections <- second
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	config := testConfig()
	config.QueueSize = 1
	manager, err := NewManager(dialer, func(ctx context.Context, _ uint, _ []byte) error {
		select {
		case started <- struct{}{}:
		default:
		}
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return context.Cause(ctx)
		}
	}, config)
	require.NoError(t, err)
	require.NoError(t, manager.Start(t.Context(), 8))
	first.reads <- []byte("one")
	<-started
	first.reads <- []byte("two")
	first.reads <- []byte("three")
	require.Eventually(t, func() bool { return dialer.calls.Load() >= 2 }, time.Second, time.Millisecond)
	close(release)
	require.True(t, manager.Stop(8))
	require.True(t, first.isClosed())
}

// TestRuntimeSnapshotRecordsReconnectError 验证对应测试场景。
func TestRuntimeSnapshotRecordsReconnectError(t *testing.T) {
	manager, err := NewManager(errorDialer{}, func(context.Context, uint, []byte) error { return nil }, testConfig())
	require.NoError(t, err)
	require.NoError(t, manager.Start(t.Context(), 11))
	require.Eventually(t, func() bool {
		snapshot, running := manager.Snapshot(11)
		return running && snapshot.State == StateReconnecting && snapshot.LastError == "连接失败"
	}, time.Second, time.Millisecond)
	require.True(t, manager.Stop(11))
}

// TestRuntimeStopsRetryingNonRetryableError 验证对应测试场景。
func TestRuntimeStopsRetryingNonRetryableError(t *testing.T) {
	dialer := &permanentErrorDialer{}
	states := make(chan State, 4)
	config := testConfig()
	config.StateChanged = func(snapshot Snapshot) { states <- snapshot.State }
	manager, err := NewManager(dialer, func(context.Context, uint, []byte) error { return nil }, config)
	require.NoError(t, err)
	require.NoError(t, manager.Start(t.Context(), 12))
	require.Eventually(t, func() bool {
		snapshot, running := manager.Snapshot(12)
		return !running && snapshot.State == StateFailed && snapshot.LastError == "需要人工处理" && snapshot.ActionURL == "https://example.com/verify"
	}, time.Second, time.Millisecond)
	require.Equal(t, int64(1), dialer.calls.Load())
	require.Equal(t, StateConnecting, <-states)
	require.Equal(t, StateFailed, <-states)
}

// TestRuntimeReadTimeoutClosesConnection 验证对应测试场景。
func TestRuntimeReadTimeoutClosesConnection(t *testing.T) {
	connection := newFakeConnection()
	config := testConfig()
	config.ReadTimeout = 10 * time.Millisecond
	runtime := &accountRuntime{accountID: 9, handler: func(context.Context, uint, []byte) error { return nil }, config: config}

	err := runtime.serve(t.Context(), connection)
	require.ErrorIs(t, err, ErrReadTimeout)
	require.True(t, connection.isClosed())
}

// TestRuntimeEmptyReadsResetTimeoutWithoutCallingHandler 验证对应测试场景。
func TestRuntimeEmptyReadsResetTimeoutWithoutCallingHandler(t *testing.T) {
	connection := newFakeConnection()
	config := testConfig()
	config.ReadTimeout = 100 * time.Millisecond
	messages := make(chan []byte, 1)
	failures := make(chan error, 1)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	runtime := &accountRuntime{accountID: 9, config: config}
	go runtime.readLoop(ctx, connection, messages, func(err error) { failures <- err })

	for range 3 {
		time.Sleep(60 * time.Millisecond)
		connection.reads <- nil
	}
	connection.reads <- []byte("message")

	select {
	case message := <-messages:
		require.Equal(t, []byte("message"), message)
	case err := <-failures:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("未收到消息")
	}
}

// TestRuntimeHeartbeatTimeoutClosesConnection 验证对应测试场景。
func TestRuntimeHeartbeatTimeoutClosesConnection(t *testing.T) {
	connection := newFakeConnection()
	connection.blockPing.Store(true)
	config := testConfig()
	runtime := &accountRuntime{accountID: 10, handler: func(context.Context, uint, []byte) error { return nil }, config: config}

	err := runtime.serve(t.Context(), connection)
	require.ErrorIs(t, err, ErrHeartbeatTimeout)
	require.True(t, connection.isClosed())
}

// TestBackoffDelayIsExponentialAndCapped 验证对应测试场景。
func TestBackoffDelayIsExponentialAndCapped(t *testing.T) {
	config := testConfig()
	config.BackoffMin = 10 * time.Millisecond
	config.BackoffMax = 40 * time.Millisecond
	config.JitterPercent = 0
	require.Equal(t, 10*time.Millisecond, backoffDelay(0, config))
	require.Equal(t, 20*time.Millisecond, backoffDelay(1, config))
	require.Equal(t, 40*time.Millisecond, backoffDelay(2, config))
	require.Equal(t, 40*time.Millisecond, backoffDelay(20, config))
	config.JitterPercent = 20
	for range 100 {
		delay := backoffDelay(20, config)
		require.GreaterOrEqual(t, delay, 32*time.Millisecond)
		require.LessOrEqual(t, delay, config.BackoffMax)
	}
}

// testConfig 提供测试辅助能力。
func testConfig() Config {
	return Config{
		QueueSize:         4,
		HeartbeatInterval: 5 * time.Millisecond,
		OperationTimeout:  20 * time.Millisecond,
		ReadTimeout:       time.Second,
		BackoffMin:        time.Millisecond,
		BackoffMax:        5 * time.Millisecond,
		BackoffResetAfter: time.Second,
	}
}

// fakeDialer 是测试辅助类型。
type fakeDialer struct {
	connections chan Connection // connections 提供模拟连接。
	calls       atomic.Int64    // calls 记录拨号次数。
}

// errorDialer 是测试辅助类型。
type errorDialer struct{}

// permanentError 是测试辅助类型。
type permanentError struct{}

// Error 提供测试辅助能力。
func (permanentError) Error() string { return "需要人工处理" }

// NonRetryable 提供测试辅助能力。
func (permanentError) NonRetryable() bool { return true }

// ActionURL 提供测试辅助能力。
func (permanentError) ActionURL() string { return "https://example.com/verify" }

// permanentErrorDialer 是测试辅助类型。
type permanentErrorDialer struct {
	calls atomic.Int64 // calls 记录拨号次数。
}

// Dial 提供测试辅助能力。
func (d *permanentErrorDialer) Dial(context.Context, uint) (Connection, error) {
	d.calls.Add(1)
	return nil, permanentError{}
}

// Dial 提供测试辅助能力。
func (errorDialer) Dial(context.Context, uint) (Connection, error) {
	return nil, errors.New("连接失败")
}

// Dial 提供测试辅助能力。
func (d *fakeDialer) Dial(ctx context.Context, _ uint) (Connection, error) {
	d.calls.Add(1)
	select {
	case connection := <-d.connections:
		return connection, nil
	case <-ctx.Done():
		return nil, context.Cause(ctx)
	}
}

// fakeConnection 是测试辅助类型。
type fakeConnection struct {
	reads     chan []byte   // reads 提供模拟入站消息。
	sends     chan []byte   // sends 记录模拟出站消息。
	closed    chan struct{} // closed 标记连接关闭。
	once      sync.Once     // once 保证连接只关闭一次。
	pings     atomic.Int64  // pings 记录心跳次数。
	blockPing atomic.Bool   // blockPing 控制心跳阻塞。
}

// newFakeConnection 提供测试辅助能力。
func newFakeConnection() *fakeConnection {
	return &fakeConnection{reads: make(chan []byte, 8), sends: make(chan []byte, 8), closed: make(chan struct{})}
}

// Read 提供测试辅助能力。
func (c *fakeConnection) Read(ctx context.Context) ([]byte, error) {
	select {
	case message := <-c.reads:
		return message, nil
	case <-c.closed:
		return nil, errors.New("连接已关闭")
	case <-ctx.Done():
		return nil, context.Cause(ctx)
	}
}

// Send 提供测试辅助能力。
func (c *fakeConnection) Send(ctx context.Context, payload []byte) error {
	select {
	case c.sends <- payload:
		return nil
	case <-c.closed:
		return errors.New("连接已关闭")
	case <-ctx.Done():
		return context.Cause(ctx)
	}
}

// Ping 提供测试辅助能力。
func (c *fakeConnection) Ping(ctx context.Context) error {
	if c.blockPing.Load() {
		<-ctx.Done()
		return context.Cause(ctx)
	}
	select {
	case <-c.closed:
		return errors.New("连接已关闭")
	case <-ctx.Done():
		return context.Cause(ctx)
	default:
		c.pings.Add(1)
		return nil
	}
}

// Close 提供测试辅助能力。
func (c *fakeConnection) Close() error {
	c.once.Do(func() { close(c.closed) })
	return nil
}

// isClosed 提供测试辅助能力。
func (c *fakeConnection) isClosed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}
