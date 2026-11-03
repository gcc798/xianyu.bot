package agent

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	agenttool "github.com/gcc798/xianyu.bot/internal/agent/tool"
	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func TestAgentHarnessEnforcesIndependentBudgets(t *testing.T) {
	config := model.DefaultAgentConfig()
	config.MaxToolCalls = 3
	config.MaxSearchCalls = 1
	config.MaxWriteToolCalls = 1
	harness := newAgentHarness(config)

	search := agenttool.Definition{Code: "resource_search", BudgetClass: agenttool.BudgetSearch}
	write := agenttool.Definition{Code: "resource_acquire", RequiresApproval: true}
	require.NoError(t, harness.BeforeToolCall(search))
	err := harness.BeforeToolCall(search)
	require.True(t, errors.Is(err, ErrAgentBudgetExceeded))
	require.NoError(t, harness.BeforeToolCall(write))
	err = harness.BeforeToolCall(write)
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
	definition := agenttool.Definition{Code: "channel_get_status"}

	var accepted atomic.Int64
	var workers sync.WaitGroup
	for range 20 {
		workers.Go(func() {
			if harness.BeforeToolCall(definition) == nil {
				accepted.Add(1)
			}
		})
	}
	workers.Wait()
	require.EqualValues(t, 5, accepted.Load())
	require.Equal(t, 5, harness.Usage().ToolCalls)
}

func TestAgentHarnessLimitsActiveToolCalls(t *testing.T) {
	config := model.DefaultAgentConfig()
	config.MaxConcurrentToolCalls = 2
	harness := newAgentHarness(config)
	require.NoError(t, harness.AcquireToolSlot(t.Context()))
	require.NoError(t, harness.AcquireToolSlot(t.Context()))

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
	defer cancel()
	require.Error(t, harness.AcquireToolSlot(ctx))
	harness.ReleaseToolSlot()
	harness.ReleaseToolSlot()
}

func TestMCPToolPoliciesClassifyUnknownToolsConservatively(t *testing.T) {
	readOnly := true
	require.False(t, mcpToolRequiresApproval(mcp.Tool{Name: "custom_read", Annotations: mcp.ToolAnnotation{ReadOnlyHint: &readOnly}}))
	require.True(t, mcpToolRequiresApproval(mcp.Tool{Name: "custom_unknown"}))
	require.Equal(t, 5*time.Second, mcpToolTimeout(mcp.Tool{Name: "resource_acquire"}))
	require.Equal(t, 35*time.Second, mcpToolTimeout(mcp.Tool{Name: "message_list"}))
	require.False(t, mcpToolRequiresApproval(mcp.Tool{Name: "message_send"}))
	require.True(t, mcpToolRequiresApproval(mcp.Tool{Name: "dialog_join"}))
	require.Equal(t, 30*time.Second, mcpToolTimeout(mcp.Tool{Name: "dialog_join"}))
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
