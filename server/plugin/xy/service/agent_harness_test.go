package service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func TestAgentHarnessEnforcesIndependentBudgets(t *testing.T) {
	config := model.DefaultAgentConfig()
	config.MaxToolCalls = 3
	config.MaxSearchCalls = 1
	config.MaxWriteToolCalls = 1
	harness := newAgentHarness(config)

	require.NoError(t, harness.BeforeToolCall(mcp.Tool{Name: "resource_search"}))
	err := harness.BeforeToolCall(mcp.Tool{Name: "resource_search"})
	require.True(t, errors.Is(err, ErrAgentBudgetExceeded))
	require.NoError(t, harness.BeforeToolCall(mcp.Tool{Name: "resource_acquire"}))
	err = harness.BeforeToolCall(mcp.Tool{Name: "task_cancel"})
	require.True(t, errors.Is(err, ErrAgentBudgetExceeded))

	usage := harness.Usage()
	require.Equal(t, 2, usage.ToolCalls)
	require.Equal(t, 1, usage.Searches)
	require.Equal(t, 1, usage.Writes)
}

func TestAgentHarnessKeepsConcurrentToolCallsWithinLimit(t *testing.T) {
	config := model.DefaultAgentConfig()
	config.MaxToolCalls = 5
	harness := newAgentHarness(config)
	tool := mcp.Tool{Name: "channel_get_status"}

	var accepted atomic.Int64
	var workers sync.WaitGroup
	for range 20 {
		workers.Go(func() {
			if harness.BeforeToolCall(tool) == nil {
				accepted.Add(1)
			}
		})
	}
	workers.Wait()
	require.EqualValues(t, 5, accepted.Load())
	require.Equal(t, 5, harness.Usage().ToolCalls)
}

func TestAgentHarnessClassifiesUnknownToolsConservatively(t *testing.T) {
	readOnly := true
	require.False(t, toolRequiresApproval(mcp.Tool{Name: "custom_read", Annotations: mcp.ToolAnnotation{ReadOnlyHint: &readOnly}}))
	require.True(t, toolRequiresApproval(mcp.Tool{Name: "custom_unknown"}))
	require.Equal(t, 5*time.Second, toolTimeout(mcp.Tool{Name: "resource_acquire"}))
	require.Equal(t, 45*time.Second, toolTimeout(mcp.Tool{Name: "resource_search"}))
}

func TestAgentHarnessAppliesWallTime(t *testing.T) {
	config := model.DefaultAgentConfig()
	config.MaxWallTimeSeconds = 1
	harness := newAgentHarness(config)
	ctx, cancel := harness.Context(t.Context())
	defer cancel()
	<-ctx.Done()
	require.ErrorIs(t, context.Cause(ctx), ErrAgentWallTime)
}
