package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/eino-contrib/jsonschema"
	agenttool "github.com/gcc798/xianyu.bot/internal/agent/tool"
	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/mark3labs/mcp-go/mcp"
	"gorm.io/gorm"
)

// mcpAgentTool 将远端 MCP 工具适配为 Agent 工具。
type mcpAgentTool struct {
	db         *gorm.DB             // db 持久化工具产生的资源候选。
	connection *mcpConnection       // connection 是工具所属 MCP 连接。
	remote     mcp.Tool             // remote 是远端工具定义。
	definition agenttool.Definition // definition 是 Agent 可见的工具定义。
}

var _ agenttool.Tool = (*mcpAgentTool)(nil)

// newMCPAgentTool 创建 MCP 工具适配器。
func newMCPAgentTool(db *gorm.DB, connection *mcpConnection, remote mcp.Tool) (*mcpAgentTool, error) {
	parameters, err := agentVisibleToolSchema(remote)
	if err != nil {
		return nil, fmt.Errorf("解析 MCP 工具 %s 参数: %w", remote.Name, err)
	}
	requiresApproval := mcpToolRequiresApproval(remote)
	budgetClass := agenttool.BudgetGeneral
	if remote.Name == "resource_search" {
		budgetClass = agenttool.BudgetSearch
	}
	timeout := mcpToolTimeout(remote)
	if serverTimeout := time.Duration(connection.Server.RequestTimeoutSeconds) * time.Second; serverTimeout > 0 {
		timeout = min(timeout, serverTimeout)
	}
	return &mcpAgentTool{
		db: db, connection: connection, remote: remote,
		definition: agenttool.Definition{
			Code: connection.Server.Code + "__" + remote.Name, Name: remote.Name,
			Description: remote.Description, Source: "mcp", ProviderCode: connection.Server.Code,
			Parameters: parameters, ReadOnly: mcpToolReadOnly(remote), RequiresApproval: requiresApproval,
			Timeout: timeout, BudgetClass: budgetClass,
		},
	}, nil
}

// Definition 返回 Agent 工具定义。
func (t *mcpAgentTool) Definition() agenttool.Definition { return t.definition }

// Invoke 调用远端 MCP 工具并保存资源候选。
func (t *mcpAgentTool) Invoke(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
	var input map[string]any
	if err := json.Unmarshal(arguments, &input); err != nil {
		return "", &agenttool.ExecutionError{Code: "invalid_arguments", Message: "工具参数无效", Cause: err}
	}
	result, err := t.connection.Client.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{
		Name: t.remote.Name, Arguments: input,
	}})
	if err != nil {
		return "", &agenttool.ExecutionError{Code: "capability_unavailable", Message: "MCP 能力暂不可用", Cause: err}
	}
	if err := t.persistResourceCandidates(ctx, scope.OwnerID, input, result); err != nil {
		return "", &agenttool.ExecutionError{Code: "metadata_persistence_failed", Message: "保存资源候选失败", Cause: err}
	}
	payload, err := json.Marshal(result)
	if err != nil {
		return "", &agenttool.ExecutionError{Code: "invalid_result", Message: "MCP 工具结果无效", Cause: err}
	}
	if result.IsError {
		return string(payload), &agenttool.ExecutionError{Code: "mcp_error", Message: "MCP 工具返回错误"}
	}
	return string(payload), nil
}

// filterArguments 移除资源领取工具的服务端受控参数。
func (t *mcpAgentTool) filterArguments(arguments json.RawMessage) (json.RawMessage, error) {
	if t.remote.Name != "resource_acquire" {
		return arguments, nil
	}
	var input map[string]any
	if err := json.Unmarshal(arguments, &input); err != nil {
		return nil, err
	}
	delete(input, "task_id")
	delete(input, "idempotency_key")
	delete(input, "delivery_token")
	return json.Marshal(input)
}

// resourceMessageFile 描述消息中的资源候选文件。
type resourceMessageFile struct {
	CandidateID string `json:"candidate_id"` // CandidateID 是外部候选标识。
	Name        string `json:"name"`         // Name 是文件名。
	Size        int64  `json:"size"`         // Size 是文件字节数。
	MIMEType    string `json:"mime_type"`    // MIMEType 是文件媒体类型。
}

// resourceMessage 描述可能携带资源文件的消息。
type resourceMessage struct {
	File *resourceMessageFile `json:"file"` // File 是资源候选文件。
	Text string               `json:"text"` // Text 是消息文本。
}

// resourceToolResult 描述资源类 MCP 工具结果。
type resourceToolResult struct {
	Status string `json:"status"` // Status 标识工具执行状态。
	Data   struct {
		Dialog   json.RawMessage   `json:"dialog"`   // Dialog 是来源会话信息。
		Messages []resourceMessage `json:"messages"` // Messages 是资源消息列表。
	} `json:"data"` // Data 承载资源结果。
}

// persistResourceCandidates 保存消息工具返回的资源候选。
func (t *mcpAgentTool) persistResourceCandidates(ctx context.Context, ownerID uint, input map[string]any, result *mcp.CallToolResult) error {
	if !slices.Contains([]string{"message_list", "message_search"}, t.remote.Name) || result == nil || result.IsError || ownerID == 0 {
		return nil
	}
	payload, err := json.Marshal(mcpToolResultValue(result))
	if err != nil {
		return err
	}
	var toolResult resourceToolResult
	if err := json.Unmarshal(payload, &toolResult); err != nil {
		return err
	}
	if toolResult.Status != "success" || !slices.ContainsFunc(toolResult.Data.Messages, func(message resourceMessage) bool { return message.File != nil }) {
		return nil
	}
	query, _ := input["query"].(string)
	expiresAt := time.Now().Add(24 * time.Hour)
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		search := model.ResourceSearch{OwnerID: ownerID, Query: query, Status: "completed"}
		if err := tx.Create(&search).Error; err != nil {
			return err
		}
		for _, message := range toolResult.Data.Messages {
			if message.File == nil {
				continue
			}
			file := message.File
			if file.CandidateID == "" || file.Name == "" || file.Size < 0 {
				return errors.New("MCP 资源候选格式无效")
			}
			source := toolResult.Data.Dialog
			if len(source) == 0 {
				source = json.RawMessage(`{}`)
			}
			candidate := model.ResourceCandidate{
				SearchID: search.ID, MCPServerCode: t.connection.Server.Code, ExternalID: file.CandidateID,
				Name: file.Name, Size: file.Size, MIMEType: file.MIMEType, Source: string(source),
				Preview: message.Text, License: "unknown", Risk: "来源未提供可验证的许可证信息", ExpiresAt: &expiresAt,
			}
			if err := tx.Create(&candidate).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// mcpToolResultValue 提取 MCP 工具的结构化结果。
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

// mcpToolRequiresApproval 判断 MCP 工具是否需要用户确认。
func mcpToolRequiresApproval(remote mcp.Tool) bool {
	switch remote.Name {
	case "channel_get_info", "channel_get_status", "dialog_list", "dialog_resolve", "message_send", "message_list", "message_search", "resource_search", "resource_inspect", "task_get":
		return false
	case "source_sync", "dialog_join", "resource_acquire", "task_cancel":
		return true
	default:
		return remote.Annotations.ReadOnlyHint == nil || !*remote.Annotations.ReadOnlyHint
	}
}

// mcpToolReadOnly 判断 MCP 工具是否只读。
func mcpToolReadOnly(remote mcp.Tool) bool {
	if remote.Annotations.ReadOnlyHint != nil {
		return *remote.Annotations.ReadOnlyHint
	}
	switch remote.Name {
	case "channel_get_info", "channel_get_status", "dialog_list", "dialog_resolve", "message_list", "message_search", "resource_search", "resource_inspect", "task_get":
		return true
	default:
		return false
	}
}

// mcpToolTimeout 返回 MCP 工具的默认执行时限。
func mcpToolTimeout(remote mcp.Tool) time.Duration {
	switch remote.Name {
	case "dialog_join":
		return 30 * time.Second
	case "message_list":
		return 35 * time.Second
	case "task_get":
		return 5 * time.Second
	}
	if mcpToolRequiresApproval(remote) {
		return 5 * time.Second
	}
	return 15 * time.Second
}

// agentVisibleToolSchema 构造 Agent 可见的工具参数结构。
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
