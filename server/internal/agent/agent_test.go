package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRunnerUsesRootContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	runner := NewRunner(ctx, nil, nil, nil, nil)
	require.Equal(t, ctx, runner.root)
}
