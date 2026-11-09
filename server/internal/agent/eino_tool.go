package agent

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	agenttool "github.com/gcc798/xianyu.bot/internal/agent/tool"
	"github.com/gcc798/xianyu.bot/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// argumentFilter 允许工具在调用前过滤模型提供的参数。
type argumentFilter interface {
	// filterArguments 返回允许传给工具的参数。
	filterArguments(json.RawMessage) (json.RawMessage, error)
}

// governedTool 为 Eino 工具调用提供预算、审批、审计和并发治理。
type governedTool struct {
	db             *gorm.DB        // db 持久化工具调用记录。
	implementation agenttool.Tool  // implementation 是实际工具实现。
	harness        *agentHarness   // harness 控制运行预算和并发。
	events         *EventHub       // events 发布工具生命周期事件。
	scope          agenttool.Scope // scope 限定工具调用主体。
	sessionID      uint            // sessionID 是 Agent 会话标识。
	runID          uint            // runID 是 Agent 运行标识。
}

// 确保 governedTool 实现 Eino 可调用工具接口。
var _ tool.InvokableTool = (*governedTool)(nil)

// newGovernedTool 创建受治理的 Eino 工具适配器。
func newGovernedTool(db *gorm.DB, implementation agenttool.Tool, harness *agentHarness, events *EventHub, run model.AgentRun) *governedTool {
	return &governedTool{
		db: db, implementation: implementation, harness: harness, events: events,
		scope: agenttool.Scope{OwnerID: run.OwnerID}, sessionID: run.SessionID, runID: run.ID,
	}
}

// Info 返回 Eino 所需的工具元数据。
func (t *governedTool) Info(context.Context) (*schema.ToolInfo, error) {
	definition := t.implementation.Definition()
	return &schema.ToolInfo{
		Name: definition.Code, Desc: definition.Description,
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(definition.Parameters),
	}, nil
}

// InvokableRun 执行一次受预算、审批和审计约束的工具调用。
func (t *governedTool) InvokableRun(ctx context.Context, arguments string, _ ...tool.Option) (string, error) {
	call, err := t.startToolCall(ctx)
	if err != nil {
		return "", err
	}
	raw := json.RawMessage(arguments)
	if filter, ok := t.implementation.(argumentFilter); ok {
		raw, err = filter.filterArguments(raw)
		if err != nil {
			return t.failBeforeExecution(ctx, call, "invalid_arguments", "工具参数无效")
		}
	}
	var input map[string]any
	if err := json.Unmarshal(raw, &input); err != nil {
		return t.failBeforeExecution(ctx, call, "invalid_arguments", "工具参数无效")
	}
	t.publish("tool.started", t.toolEvent(call, map[string]any{"arguments": input}))
	definition := t.implementation.Definition()
	if err := t.harness.BeforeToolCall(definition); err != nil {
		return t.failBeforeExecution(ctx, call, "budget_exceeded", "工具调用次数已达到限制")
	}
	if definition.RequiresApproval {
		if err := t.queueApproval(ctx, call, raw); err != nil {
			return "", err
		}
		t.publish("approval.required", t.toolEvent(call, map[string]any{"arguments": input}))
		return toolFailure("approval_required", "此操作需要用户确认后执行"), nil
	}
	if err := t.saveToolArguments(ctx, call, raw); err != nil {
		return "", err
	}
	callContext, cancel := context.WithTimeout(ctx, definition.Timeout)
	defer cancel()
	if err := t.harness.AcquireToolSlot(callContext); err != nil {
		return t.failBeforeExecution(ctx, call, "capability_unavailable", "等待工具执行超时")
	}
	defer t.harness.ReleaseToolSlot()
	output, invokeErr := t.implementation.Invoke(callContext, t.scope, raw)
	status, errorCode, errorMessage := "completed", "", ""
	if invokeErr != nil {
		status = "failed"
		if executionErr, ok := errors.AsType[*agenttool.ExecutionError](invokeErr); ok {
			errorCode, errorMessage = executionErr.Code, executionErr.Message
		} else {
			errorCode, errorMessage = "capability_unavailable", "工具能力暂不可用"
		}
		if output == "" {
			output = toolFailure(errorCode, errorMessage)
		}
	}
	if err := t.saveToolResult(ctx, call, output); err != nil {
		return "", err
	}
	if err := t.finishToolCall(ctx, call, status, errorCode, errorMessage); err != nil {
		return "", err
	}
	eventData := map[string]any{
		"status": status, "trace": compactToolOutput(output),
		"resultRef": map[string]any{"type": "agent_run_tool", "id": toolCallRecordID(call)},
	}
	if status == "failed" {
		eventData["error"] = map[string]any{"code": errorCode, "message": errorMessage}
		t.publish("tool.failed", t.toolEvent(call, eventData))
	} else {
		t.publish("tool.completed", t.toolEvent(call, eventData))
	}
	return output, nil
}

// failBeforeExecution 记录未进入实际执行阶段的工具失败。
func (t *governedTool) failBeforeExecution(ctx context.Context, call *model.AgentRunTool, code, message string) (string, error) {
	if err := t.finishToolCall(ctx, call, "failed", code, message); err != nil {
		return "", err
	}
	t.publish("tool.failed", t.toolEvent(call, map[string]any{"error": map[string]any{"code": code, "message": message}}))
	return toolFailure(code, message), nil
}

// queueApproval 将工具调用置为等待用户确认。
func (t *governedTool) queueApproval(ctx context.Context, call *model.AgentRunTool, arguments json.RawMessage) error {
	if call == nil {
		return nil
	}
	completedAt := time.Now()
	finishContext, cancel := agentFinishContext(ctx)
	defer cancel()
	return t.db.WithContext(finishContext).Model(call).Updates(map[string]any{
		"arguments": string(arguments), "status": "waiting_approval", "error_code": "approval_required",
		"error_message": "等待用户确认", "completed_at": completedAt,
		"duration_millis": time.Since(call.StartedAt).Milliseconds(),
	}).Error
}

// startToolCall 创建工具调用记录。
func (t *governedTool) startToolCall(ctx context.Context) (*model.AgentRunTool, error) {
	if t.runID == 0 {
		return nil, nil
	}
	callID := compose.GetToolCallID(ctx)
	if callID == "" {
		callID = fmt.Sprintf("run-%d-%s", t.runID, rand.Text())
	}
	definition := t.implementation.Definition()
	call := &model.AgentRunTool{
		RunID: t.runID, CallID: callID, MCPServerCode: definition.ProviderCode,
		ToolCode: definition.Code, ToolName: definition.Name, ToolSource: definition.Source,
		Status: "running", StartedAt: time.Now(),
	}
	if err := t.db.WithContext(ctx).Create(call).Error; err != nil {
		return nil, err
	}
	return call, nil
}

// saveToolArguments 保存实际提交的工具参数。
func (t *governedTool) saveToolArguments(ctx context.Context, call *model.AgentRunTool, arguments json.RawMessage) error {
	if call == nil {
		return nil
	}
	return t.db.WithContext(ctx).Model(call).Update("arguments", string(arguments)).Error
}

// saveToolResult 保存工具输出。
func (t *governedTool) saveToolResult(ctx context.Context, call *model.AgentRunTool, output string) error {
	if call == nil {
		return nil
	}
	result := any(output)
	if json.Valid([]byte(output)) {
		result = json.RawMessage(output)
	}
	payload, err := json.Marshal(result)
	if err != nil {
		return err
	}
	finishContext, cancel := agentFinishContext(ctx)
	defer cancel()
	return t.db.WithContext(finishContext).Model(call).Update("result", string(payload)).Error
}

// compactToolOutput 压缩事件中的工具输出，避免事件过大。
func compactToolOutput(output string) any {
	var value any
	if json.Unmarshal([]byte(output), &value) != nil {
		value = output
	}
	compacted := compactTraceValue(value)
	payload, err := json.Marshal(compacted)
	if err == nil && len(payload) <= agentEventMaxBytes/2 {
		return compacted
	}
	return map[string]any{"truncated": true, "message": "工具结果过大，请根据游标分批读取"}
}

// compactTraceValue 递归限制轨迹中的集合和文本长度。
func compactTraceValue(value any) any {
	switch current := value.(type) {
	case map[string]any:
		compacted := make(map[string]any, len(current))
		for key, nested := range current {
			compacted[key] = compactTraceValue(nested)
		}
		return compacted
	case []any:
		compacted := make([]any, 0, min(len(current), 100))
		for _, nested := range current[:min(len(current), 100)] {
			compacted = append(compacted, compactTraceValue(nested))
		}
		return compacted
	case string:
		runes := []rune(current)
		if len(runes) > 500 {
			return string(runes[:500]) + "…"
		}
		return current
	default:
		return value
	}
}

// finishToolCall 持久化工具调用的最终状态。
func (t *governedTool) finishToolCall(ctx context.Context, call *model.AgentRunTool, status, errorCode, errorMessage string) error {
	if call == nil {
		return nil
	}
	completedAt := time.Now()
	finishContext, cancel := agentFinishContext(ctx)
	defer cancel()
	return t.db.WithContext(finishContext).Model(call).Updates(map[string]any{
		"status": status, "duration_millis": time.Since(call.StartedAt).Milliseconds(),
		"error_code": errorCode, "error_message": errorMessage, "completed_at": completedAt,
	}).Error
}

// publish 发布工具生命周期事件。
func (t *governedTool) publish(eventType string, data any) {
	if t.events != nil {
		if _, err := t.events.Publish(t.sessionID, eventType, data); err != nil {
			zap.L().Error("保存 Agent 工具事件失败", zap.Uint("sessionID", t.sessionID), zap.String("eventType", eventType), zap.Error(err))
		}
	}
}

// toolEvent 补齐工具事件的运行和调用上下文。
func (t *governedTool) toolEvent(call *model.AgentRunTool, data map[string]any) map[string]any {
	definition := t.implementation.Definition()
	data["runId"] = t.runID
	data["callId"] = toolCallID(call)
	data["server"] = definition.ProviderCode
	data["tool"] = definition.Name
	data["toolCode"] = definition.Code
	data["toolName"] = definition.Name
	if call != nil {
		data["durationMillis"] = time.Since(call.StartedAt).Milliseconds()
	}
	return data
}

// toolCallID 返回工具调用业务标识。
func toolCallID(call *model.AgentRunTool) string {
	if call == nil {
		return ""
	}
	return call.CallID
}

// toolCallRecordID 返回工具调用记录标识。
func toolCallRecordID(call *model.AgentRunTool) uint {
	if call == nil {
		return 0
	}
	return call.ID
}

// toolFailure 序列化模型可识别的工具失败结果。
func toolFailure(code, message string) string {
	payload, _ := json.Marshal(map[string]any{
		"ok": false, "error": map[string]string{"code": code, "message": message},
	})
	return string(payload)
}
