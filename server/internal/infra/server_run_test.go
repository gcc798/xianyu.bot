package infra

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

// testServer 模拟可控的 HTTP 服务生命周期。
type testServer struct {
	listenErr error         // listenErr 是启动时应返回的错误。
	stopped   chan struct{} // stopped 在模拟服务关闭时解除阻塞。
	shutdown  atomic.Bool   // shutdown 记录是否调用过关闭流程。
}

// ListenAndServe 模拟阻塞监听或立即返回配置的错误。
func (s *testServer) ListenAndServe() error {
	if s.listenErr != nil {
		return s.listenErr
	}
	<-s.stopped
	return http.ErrServerClosed
}

// Shutdown 记录调用并停止模拟监听。
func (s *testServer) Shutdown(context.Context) error {
	s.shutdown.Store(true)
	close(s.stopped)
	return nil
}

// TestRunServerReturnsListenError 验证监听失败会原样进入错误链。
func TestRunServerReturnsListenError(t *testing.T) {
	expected := errors.New("address already in use")
	err := runServer(t.Context(), &testServer{listenErr: expected})
	require.ErrorIs(t, err, expected)
}

// TestRunServerShutsDownAfterCancellation 验证上下文取消会触发优雅关闭。
func TestRunServerShutsDownAfterCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	srv := &testServer{stopped: make(chan struct{})}

	require.NoError(t, runServer(ctx, srv))
	require.True(t, srv.shutdown.Load())
}
