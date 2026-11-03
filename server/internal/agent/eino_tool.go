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

type argumentFilter interface {
	filterArguments(json.RawMessage) (json.RawMessage, error)
}

type governedTool struct {
	db             *gorm.DB
	implementation agenttool.Tool
	harness        *agentHarness
	events         *EventHub
	scope          agenttool.Scope
	sessionID      uint
	runID          uint
}

var _ tool.InvokableTool = (*governedTool)(nil)

func newGovernedTool(db *gorm.DB, implementation agenttool.Tool, harness *agentHarness, events *EventHub, run model.AgentRun) *governedTool {
	return &governedTool{
		db: db, implementation: implementation, harness: harness, events: events,
		scope: agenttool.Scope{OwnerID: run.OwnerID}, sessionID: run.SessionID, runID: run.ID,
	}
}

func (t *governedTool) Info(context.Context) (*schema.ToolInfo, error) {
	definition := t.implementation.Definition()
	return &schema.ToolInfo{
		Name: definition.Code, Desc: definition.Description,
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(definition.Parameters),
	}, nil
}

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

func (t *governedTool) failBeforeExecution(ctx context.Context, call *model.AgentRunTool, code, message string) (string, error) {
	if err := t.finishToolCall(ctx, call, "failed", code, message); err != nil {
		return "", err
	}
	t.publish("tool.failed", t.toolEvent(call, map[string]any{"error": map[string]any{"code": code, "message": message}}))
	return toolFailure(code, message), nil
}

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

func (t *governedTool) saveToolArguments(ctx context.Context, call *model.AgentRunTool, arguments json.RawMessage) error {
	if call == nil {
		return nil
	}
	return t.db.WithContext(ctx).Model(call).Update("arguments", string(arguments)).Error
}

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

func (t *governedTool) publish(eventType string, data any) {
	if t.events != nil {
		if _, err := t.events.Publish(t.sessionID, eventType, data); err != nil {
			zap.L().Error("保存 Agent 工具事件失败", zap.Uint("sessionID", t.sessionID), zap.String("eventType", eventType), zap.Error(err))
		}
	}
}

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

func toolCallID(call *model.AgentRunTool) string {
	if call == nil {
		return ""
	}
	return call.CallID
}

func toolCallRecordID(call *model.AgentRunTool) uint {
	if call == nil {
		return 0
	}
	return call.ID
}

func toolFailure(code, message string) string {
	payload, _ := json.Marshal(map[string]any{
		"ok": false, "error": map[string]string{"code": code, "message": message},
	})
	return string(payload)
}
