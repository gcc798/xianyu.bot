package service

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
	"github.com/mark3labs/mcp-go/mcp"
)

type einoMCPTool struct {
	connection *mcpConnection
	definition mcp.Tool
	alias      string
	harness    *agentHarness
	parameters *jsonschema.Schema
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
	if err := t.harness.BeforeToolCall(t.definition); err != nil {
		return toolFailure("budget_exceeded", "本轮工具调用次数已达到限制"), nil
	}
	if toolRequiresApproval(t.definition) {
		return toolFailure("approval_required", "此操作需要用户确认后执行"), nil
	}
	var input map[string]any
	if err := json.Unmarshal([]byte(arguments), &input); err != nil {
		return toolFailure("invalid_arguments", "工具参数无效"), nil
	}
	callContext, cancel := context.WithTimeout(ctx, min(toolTimeout(t.definition), time.Duration(t.connection.Server.RequestTimeoutSeconds)*time.Second))
	defer cancel()
	result, err := t.connection.Client.CallTool(callContext, mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: t.definition.Name, Arguments: input,
	}})
	if err != nil {
		return toolFailure("capability_unavailable", "MCP 能力暂不可用"), nil
	}
	payload, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("编码 MCP 工具结果: %w", err)
	}
	return string(payload), nil
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
