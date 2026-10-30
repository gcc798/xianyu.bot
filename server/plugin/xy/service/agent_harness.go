package service

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model"
	"github.com/mark3labs/mcp-go/mcp"
)

var (
	ErrAgentBudgetExceeded = errors.New("agent 运行预算已用尽")
	ErrAgentWallTime       = errors.New("agent 运行超时")
)

type agentLimitError struct {
	name  string
	limit int64
}

func (e *agentLimitError) Error() string {
	return fmt.Sprintf("%s 已达到上限 %d: %v", e.name, e.limit, ErrAgentBudgetExceeded)
}

func (e *agentLimitError) Unwrap() error { return ErrAgentBudgetExceeded }

type agentHarness struct {
	config     model.AgentConfig
	toolSlots  chan struct{}
	iterations atomic.Int64
	modelCalls atomic.Int64
	toolCalls  atomic.Int64
	searches   atomic.Int64
	writes     atomic.Int64
}

type agentUsage struct {
	Iterations int
	ModelCalls int
	ToolCalls  int
	Searches   int
	Writes     int
}

func newAgentHarness(config model.AgentConfig) *agentHarness {
	return &agentHarness{config: config, toolSlots: make(chan struct{}, max(config.MaxConcurrentToolCalls, 1))}
}

func (h *agentHarness) AcquireToolSlot(ctx context.Context) error {
	select {
	case h.toolSlots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return context.Cause(ctx)
	}
}

func (h *agentHarness) ReleaseToolSlot() { <-h.toolSlots }

func (h *agentHarness) Context(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeoutCause(parent, time.Duration(h.config.MaxWallTimeSeconds)*time.Second, ErrAgentWallTime)
}

func (h *agentHarness) BeforeIteration() error {
	return reserveAgentBudget(&h.iterations, int64(h.config.MaxIterations), "Agent 迭代次数")
}

func (h *agentHarness) BeforeModelCall() error {
	return reserveAgentBudget(&h.modelCalls, int64(h.config.MaxModelCalls), "模型调用次数")
}

func (h *agentHarness) BeforeToolCall(tool mcp.Tool) error {
	if err := reserveAgentBudget(&h.toolCalls, int64(h.config.MaxToolCalls), "工具调用次数"); err != nil {
		return err
	}
	if tool.Name == "resource_search" {
		if err := reserveAgentBudget(&h.searches, int64(h.config.MaxSearchCalls), "资源检索次数"); err != nil {
			h.toolCalls.Add(-1)
			return err
		}
	}
	if toolRequiresApproval(tool) {
		if err := reserveAgentBudget(&h.writes, int64(h.config.MaxWriteToolCalls), "写工具调用次数"); err != nil {
			h.toolCalls.Add(-1)
			return err
		}
	}
	return nil
}

func (h *agentHarness) Usage() agentUsage {
	return agentUsage{
		Iterations: int(h.iterations.Load()), ModelCalls: int(h.modelCalls.Load()),
		ToolCalls: int(h.toolCalls.Load()), Searches: int(h.searches.Load()), Writes: int(h.writes.Load()),
	}
}

func reserveAgentBudget(counter *atomic.Int64, limit int64, name string) error {
	for {
		current := counter.Load()
		if current >= limit {
			return &agentLimitError{name: name, limit: limit}
		}
		if counter.CompareAndSwap(current, current+1) {
			return nil
		}
	}
}

func toolRequiresApproval(tool mcp.Tool) bool {
	switch tool.Name {
	case "channel_get_info", "channel_get_status", "dialog_list", "dialog_resolve", "message_send", "message_list", "message_search", "resource_search", "resource_inspect", "task_get":
		return false
	case "source_sync", "dialog_join", "resource_acquire", "task_cancel":
		return true
	default:
		return tool.Annotations.ReadOnlyHint == nil || !*tool.Annotations.ReadOnlyHint
	}
}

func toolTimeout(tool mcp.Tool) time.Duration {
	if tool.Name == "dialog_join" {
		return 30 * time.Second
	}
	if tool.Name == "message_list" {
		return 35 * time.Second
	}
	if toolRequiresApproval(tool) || tool.Name == "task_get" {
		return 5 * time.Second
	}
	return 15 * time.Second
}
