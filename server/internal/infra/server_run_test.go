package infra

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

type testServer struct {
	listenErr error
	stopped   chan struct{}
	shutdown  atomic.Bool
}

func (s *testServer) ListenAndServe() error {
	if s.listenErr != nil {
		return s.listenErr
	}
	<-s.stopped
	return http.ErrServerClosed
}

func (s *testServer) Shutdown(context.Context) error {
	s.shutdown.Store(true)
	close(s.stopped)
	return nil
}

func TestRunServerReturnsListenError(t *testing.T) {
	expected := errors.New("address already in use")
	err := runServer(t.Context(), &testServer{listenErr: expected})
	require.ErrorIs(t, err, expected)
}

func TestRunServerShutsDownAfterCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	srv := &testServer{stopped: make(chan struct{})}

	require.NoError(t, runServer(ctx, srv))
	require.True(t, srv.shutdown.Load())
}
