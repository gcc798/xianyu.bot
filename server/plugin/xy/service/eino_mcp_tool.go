package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model"
	"github.com/mark3labs/mcp-go/mcp"
	"gorm.io/gorm"
)

type einoMCPTool struct {
	connection *mcpConnection
	definition mcp.Tool
	alias      string
	harness    *agentHarness
	parameters *jsonschema.Schema
	events     *agentEventHub
	ownerID    uint
	sessionID  uint
	runID      uint
}

var _ tool.InvokableTool = (*einoMCPTool)(nil)

func newEinoMCPTool(connection *mcpConnection, definition mcp.Tool, harness *agentHarness) (*einoMCPTool, error) {
	parameters, err := agentVisibleToolSchema(definition)
	if err != nil {
		return nil, fmt.Errorf("解析 MCP 工具 %s 参数: %w", definition.Name, err)
	}
	return &einoMCPTool{
		connection: connection, definition: definition,
		alias:   connection.Server.Code + "__" + definition.Name,
		harness: harness, parameters: parameters,
	}, nil
}

func (t *einoMCPTool) Info(context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.alias, Desc: t.definition.Description,
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(t.parameters),
	}, nil
}

func (t *einoMCPTool) InvokableRun(ctx context.Context, arguments string, _ ...tool.Option) (string, error) {
	call, err := t.startToolCall(ctx)
	if err != nil {
		return "", err
	}
	t.publish("tool.started", map[string]any{"runId": t.runID, "tool": t.definition.Name})
	var input map[string]any
	if err := json.Unmarshal([]byte(arguments), &input); err != nil {
		if err := t.finishToolCall(ctx, call, "failed", "invalid_arguments", "工具参数无效"); err != nil {
			return "", err
		}
		t.publish("tool.completed", map[string]any{"runId": t.runID, "tool": t.definition.Name, "status": "failed"})
		return toolFailure("invalid_arguments", "工具参数无效"), nil
	}
	if err := t.harness.BeforeToolCall(t.definition); err != nil {
		if err := t.finishToolCall(ctx, call, "failed", "budget_exceeded", "工具调用次数已达到限制"); err != nil {
			return "", err
		}
		t.publish("tool.completed", map[string]any{"runId": t.runID, "tool": t.definition.Name, "status": "failed"})
		return toolFailure("budget_exceeded", "本轮工具调用次数已达到限制"), nil
	}
	if toolRequiresApproval(t.definition) {
		if err := t.queueApproval(ctx, call, input); err != nil {
			return "", err
		}
		t.publish("approval.required", map[string]any{"runId": t.runID, "tool": t.definition.Name})
		return toolFailure("approval_required", "此操作需要用户确认后执行"), nil
	}
	callContext, cancel := context.WithTimeout(ctx, min(toolTimeout(t.definition), time.Duration(t.connection.Server.RequestTimeoutSeconds)*time.Second))
	defer cancel()
	result, err := t.connection.Client.CallTool(callContext, mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: t.definition.Name, Arguments: input,
	}})
	if err != nil {
		if err := t.finishToolCall(ctx, call, "failed", "capability_unavailable", "MCP 能力暂不可用"); err != nil {
			return "", err
		}
		t.publish("tool.completed", map[string]any{"runId": t.runID, "tool": t.definition.Name, "status": "failed"})
		return toolFailure("capability_unavailable", "MCP 能力暂不可用"), nil
	}
	if err := t.persistResourceSearch(ctx, input, result); err != nil {
		if finishErr := t.finishToolCall(ctx, call, "failed", "metadata_persistence_failed", "保存资源候选失败"); finishErr != nil {
			return "", errors.Join(err, finishErr)
		}
		t.publish("tool.completed", map[string]any{"runId": t.runID, "tool": t.definition.Name, "status": "failed"})
		return toolFailure("metadata_persistence_failed", "资源候选暂时无法保存"), nil
	}
	payload, err := json.Marshal(result)
	if err != nil {
		if finishErr := t.finishToolCall(ctx, call, "failed", "invalid_result", "MCP 工具结果无效"); finishErr != nil {
			return "", errors.Join(err, finishErr)
		}
		return "", fmt.Errorf("编码 MCP 工具结果: %w", err)
	}
	status, errorCode := "completed", ""
	if result.IsError {
		status, errorCode = "failed", "mcp_error"
	}
	if err := t.finishToolCall(ctx, call, status, errorCode, ""); err != nil {
		return "", err
	}
	t.publish("tool.completed", map[string]any{"runId": t.runID, "tool": t.definition.Name, "status": status})
	return string(payload), nil
}

func (t *einoMCPTool) queueApproval(ctx context.Context, call *model.AgentRunTool, input map[string]any) error {
	if call == nil {
		return nil
	}
	if t.definition.Name == "resource_acquire" {
		delete(input, "task_id")
		delete(input, "idempotency_key")
		delete(input, "delivery_token")
	}
	arguments, err := json.Marshal(input)
	if err != nil {
		return err
	}
	completedAt := time.Now()
	finishContext, cancel := agentFinishContext(ctx)
	defer cancel()
	return global.GVA_DB.WithContext(finishContext).Model(call).Updates(map[string]any{
		"arguments": string(arguments), "status": "waiting_approval", "error_code": "approval_required",
		"error_message": "等待用户确认", "completed_at": completedAt,
		"duration_millis": time.Since(call.StartedAt).Milliseconds(),
	}).Error
}

type resourceSearchResult struct {
	Status string `json:"status"`
	Data   struct {
		Candidates []struct {
			CandidateID string          `json:"candidate_id"`
			Name        string          `json:"name"`
			Size        int64           `json:"size"`
			MIMEType    string          `json:"mime_type"`
			Source      json.RawMessage `json:"source"`
			Preview     string          `json:"preview"`
			License     string          `json:"license"`
			Risk        string          `json:"risk"`
		} `json:"candidates"`
	} `json:"data"`
}

func (t *einoMCPTool) persistResourceSearch(ctx context.Context, input map[string]any, result *mcp.CallToolResult) error {
	if t.definition.Name != "resource_search" || result == nil || result.IsError || t.ownerID == 0 {
		return nil
	}
	payload, err := json.Marshal(mcpToolResultValue(result))
	if err != nil {
		return err
	}
	var searchResult resourceSearchResult
	if err := json.Unmarshal(payload, &searchResult); err != nil {
		return err
	}
	if searchResult.Status != "success" {
		return nil
	}
	query, _ := input["query"].(string)
	expiresAt := time.Now().Add(24 * time.Hour)
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		search := model.ResourceSearch{OwnerID: t.ownerID, Query: query, Status: "completed"}
		if err := tx.Create(&search).Error; err != nil {
			return err
		}
		for _, item := range searchResult.Data.Candidates {
			if item.CandidateID == "" || item.Name == "" || item.Size < 0 {
				return errors.New("MCP 资源候选格式无效")
			}
			license := item.License
			if license == "" {
				license = "unknown"
			}
			source := item.Source
			if len(source) == 0 {
				source = json.RawMessage(`{}`)
			}
			candidate := model.ResourceCandidate{
				SearchID: search.ID, MCPServerCode: t.connection.Server.Code, ExternalID: item.CandidateID,
				Name: item.Name, Size: item.Size, MIMEType: item.MIMEType, Source: string(source),
				Preview: item.Preview, License: license, Risk: item.Risk, ExpiresAt: &expiresAt,
			}
			if err := tx.Create(&candidate).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func mcpToolResultValue(result *mcp.CallToolResult) any {
	if result == nil {
		return nil
	}
	if result.StructuredContent != nil {
		return result.StructuredContent
	}
	for _, content := range result.Content {
		var text string
		switch value := content.(type) {
		case mcp.TextContent:
			text = value.Text
		case *mcp.TextContent:
			text = value.Text
		}
		if text == "" {
			continue
		}
		var value any
		if json.Unmarshal([]byte(text), &value) == nil {
			return value
		}
	}
	return nil
}

func (t *einoMCPTool) startToolCall(ctx context.Context) (*model.AgentRunTool, error) {
	if t.runID == 0 {
		return nil, nil
	}
	callID := compose.GetToolCallID(ctx)
	if callID == "" {
		callID = fmt.Sprintf("run-%d-%s", t.runID, rand.Text())
	}
	call := &model.AgentRunTool{
		RunID: t.runID, CallID: callID, MCPServerCode: t.connection.Server.Code,
		ToolName: t.definition.Name, Status: "running", StartedAt: time.Now(),
	}
	if err := global.GVA_DB.WithContext(ctx).Create(call).Error; err != nil {
		return nil, err
	}
	return call, nil
}

func (t *einoMCPTool) finishToolCall(ctx context.Context, call *model.AgentRunTool, status, errorCode, errorMessage string) error {
	if call == nil {
		return nil
	}
	completedAt := time.Now()
	finishContext, cancel := agentFinishContext(ctx)
	defer cancel()
	return global.GVA_DB.WithContext(finishContext).Model(call).Updates(map[string]any{
		"status": status, "duration_millis": time.Since(call.StartedAt).Milliseconds(),
		"error_code": errorCode, "error_message": errorMessage, "completed_at": completedAt,
	}).Error
}

func (t *einoMCPTool) publish(eventType string, data any) {
	if t.events != nil {
		_, _ = t.events.Publish(t.sessionID, eventType, data)
	}
}

func agentVisibleToolSchema(definition mcp.Tool) (*jsonschema.Schema, error) {
	var raw []byte
	var err error
	if definition.RawInputSchema != nil {
		raw = definition.RawInputSchema
	} else {
		raw, err = json.Marshal(definition.InputSchema)
		if err != nil {
			return nil, err
		}
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	if definition.Name == "resource_acquire" {
		if properties, ok := document["properties"].(map[string]any); ok {
			delete(properties, "task_id")
			delete(properties, "idempotency_key")
			delete(properties, "delivery_token")
		}
		if required, ok := document["required"].([]any); ok {
			document["required"] = slices.DeleteFunc(required, func(value any) bool {
				name, _ := value.(string)
				return name == "task_id" || name == "idempotency_key" || name == "delivery_token"
			})
		}
	}
	raw, err = json.Marshal(document)
	if err != nil {
		return nil, err
	}
	parameters := new(jsonschema.Schema)
	if err := json.Unmarshal(raw, parameters); err != nil {
		return nil, err
	}
	return parameters, nil
}

func toolFailure(code, message string) string {
	payload, _ := json.Marshal(map[string]any{
		"ok": false, "error": map[string]string{"code": code, "message": message},
	})
	return string(payload)
}
