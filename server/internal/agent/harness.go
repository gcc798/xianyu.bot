package agent

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	agenttool "github.com/gcc798/xianyu.bot/internal/agent/tool"
	"github.com/gcc798/xianyu.bot/internal/model"
)

var (
	// ErrAgentBudgetExceeded 表示 Agent 已耗尽调用预算。
	ErrAgentBudgetExceeded = errors.New("agent 运行预算已用尽")
	// ErrAgentWallTime 表示 Agent 已超过最长运行时间。
	ErrAgentWallTime = errors.New("agent 运行超时")
)

// agentLimitError 描述具体的 Agent 预算超限项。
type agentLimitError struct {
	name  string // name 是预算项名称。
	limit int64  // limit 是预算上限。
}

// Error 返回预算超限说明。
func (e *agentLimitError) Error() string {
	return fmt.Sprintf("%s 已达到上限 %d: %v", e.name, e.limit, ErrAgentBudgetExceeded)
}

// Unwrap 返回统一的预算耗尽错误。
func (e *agentLimitError) Unwrap() error { return ErrAgentBudgetExceeded }

// agentHarness 控制单次 Agent 运行的时间、并发和调用预算。
type agentHarness struct {
	config     model.AgentConfig // config 是本次运行配置。
	toolSlots  chan struct{}     // toolSlots 限制工具并发数。
	iterations atomic.Int64      // iterations 记录迭代次数。
	modelCalls atomic.Int64      // modelCalls 记录模型调用次数。
	toolCalls  atomic.Int64      // toolCalls 记录工具调用次数。
	searches   atomic.Int64      // searches 记录检索调用次数。
	writes     atomic.Int64      // writes 记录写工具调用次数。
}

// agentUsage 汇总 Agent 运行期间的预算用量。
type agentUsage struct {
	Iterations int // Iterations 是迭代次数。
	ModelCalls int // ModelCalls 是模型调用次数。
	ToolCalls  int // ToolCalls 是工具调用次数。
	Searches   int // Searches 是检索调用次数。
	Writes     int // Writes 是写工具调用次数。
}

// newAgentHarness 创建 Agent 运行预算控制器。
func newAgentHarness(config model.AgentConfig) *agentHarness {
	return &agentHarness{config: config, toolSlots: make(chan struct{}, max(config.MaxConcurrentToolCalls, 1))}
}

// AcquireToolSlot 等待并占用一个工具并发槽位。
func (h *agentHarness) AcquireToolSlot(ctx context.Context) error {
	select {
	case h.toolSlots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return context.Cause(ctx)
	}
}

// ReleaseToolSlot 释放一个工具并发槽位。
func (h *agentHarness) ReleaseToolSlot() { <-h.toolSlots }

// Context 创建带最长运行时间限制的上下文。
func (h *agentHarness) Context(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeoutCause(parent, time.Duration(h.config.MaxWallTimeSeconds)*time.Second, ErrAgentWallTime)
}

// BeforeIteration 为一次 Agent 迭代预留预算。
func (h *agentHarness) BeforeIteration() error {
	return reserveAgentBudget(&h.iterations, int64(h.config.MaxIterations), "Agent 迭代次数")
}

// BeforeModelCall 为一次模型调用预留预算。
func (h *agentHarness) BeforeModelCall() error {
	return reserveAgentBudget(&h.modelCalls, int64(h.config.MaxModelCalls), "模型调用次数")
}

// BeforeToolCall 按工具类型预留调用预算。
func (h *agentHarness) BeforeToolCall(definition agenttool.Definition) error {
	if err := reserveAgentBudget(&h.toolCalls, int64(h.config.MaxToolCalls), "工具调用次数"); err != nil {
		return err
	}
	if definition.BudgetClass == agenttool.BudgetSearch {
		if err := reserveAgentBudget(&h.searches, int64(h.config.MaxSearchCalls), "资源检索次数"); err != nil {
			h.toolCalls.Add(-1)
			return err
		}
	}
	if definition.RequiresApproval {
		if err := reserveAgentBudget(&h.writes, int64(h.config.MaxWriteToolCalls), "写工具调用次数"); err != nil {
			h.toolCalls.Add(-1)
			return err
		}
	}
	return nil
}

// Usage 返回当前预算用量。
func (h *agentHarness) Usage() agentUsage {
	return agentUsage{
		Iterations: int(h.iterations.Load()), ModelCalls: int(h.modelCalls.Load()),
		ToolCalls: int(h.toolCalls.Load()), Searches: int(h.searches.Load()), Writes: int(h.writes.Load()),
	}
}

// reserveAgentBudget 原子预留一个预算名额。
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
