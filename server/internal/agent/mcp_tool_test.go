package agent

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"

	agenttool "github.com/gcc798/xianyu.bot/internal/agent/tool"
	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/glebarez/sqlite"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestEinoMCPToolHidesDeliveryFieldsAndRequiresApproval 验证对应测试场景。
func TestEinoMCPToolHidesDeliveryFieldsAndRequiresApproval(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentRunTool{}))
	client := new(recordingMCPClient)
	connection := &mcpConnection{
		Server: model.MCPServer{Code: "telegram", RequestTimeoutSeconds: 15}, Client: client,
	}
	definition := mcp.Tool{
		Name: "resource_acquire",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"candidate_id": map[string]any{"type": "string"},
				"task_id":      map[string]any{"type": "integer"}, "idempotency_key": map[string]any{"type": "string"},
				"delivery_token": map[string]any{"type": "string"},
			},
			Required: []string{"candidate_id", "task_id", "idempotency_key", "delivery_token"},
		},
	}
	target, err := newTestMCPTool(db, connection, definition)
	require.NoError(t, err)
	info, err := target.Info(t.Context())
	require.NoError(t, err)
	visibleSchema, err := info.ParamsOneOf.ToJSONSchema()
	require.NoError(t, err)
	payload, err := json.Marshal(visibleSchema)
	require.NoError(t, err)
	require.Contains(t, string(payload), "candidate_id")
	require.NotContains(t, string(payload), "task_id")
	require.NotContains(t, string(payload), "idempotency_key")
	require.NotContains(t, string(payload), "delivery_token")

	target.runID = 9
	result, err := target.InvokableRun(t.Context(), `{"candidate_id":"candidate-1","delivery_token":"must-not-persist"}`)
	require.NoError(t, err)
	require.Contains(t, result, "approval_required")
	require.Zero(t, client.calls.Load())
	var call model.AgentRunTool
	require.NoError(t, db.First(&call).Error)
	require.Equal(t, `{"candidate_id":"candidate-1"}`, call.Arguments)
	require.NotContains(t, call.Arguments, "must-not-persist")
}

// TestEinoMCPToolCallsReadCapability 验证对应测试场景。
func TestEinoMCPToolCallsReadCapability(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentSession{}, &model.AgentRunTool{}, &model.AgentEvent{}))
	client := new(recordingMCPClient)
	connection := &mcpConnection{
		Server: model.MCPServer{Code: "telegram", RequestTimeoutSeconds: 15}, Client: client,
	}
	definition := mcp.Tool{
		Name: "message_search",
		InputSchema: mcp.ToolInputSchema{Type: "object", Properties: map[string]any{
			"query": map[string]any{"type": "string"},
		}},
	}
	target, err := newTestMCPTool(db, connection, definition)
	require.NoError(t, err)
	session := model.AgentSession{OwnerID: 11, Title: "工具轨迹"}
	require.NoError(t, db.Create(&session).Error)
	target.runID = 9
	target.sessionID = session.ID
	target.events = NewEventHub(db)
	result, err := target.InvokableRun(t.Context(), `{"query":"Go"}`)
	require.NoError(t, err)
	require.Contains(t, result, "structuredContent")
	require.EqualValues(t, 1, client.calls.Load())
	require.Equal(t, "message_search", client.lastName)
	var call model.AgentRunTool
	require.NoError(t, db.First(&call).Error)
	require.EqualValues(t, 9, call.RunID)
	require.Equal(t, `{"query":"Go"}`, call.Arguments)
	require.Equal(t, "completed", call.Status)
	require.Empty(t, call.ErrorMessage)
	require.JSONEq(t, `{"content":[],"structuredContent":{"ok":true}}`, call.Result)
	var events []model.AgentEvent
	require.NoError(t, db.Where("session_id = ?", session.ID).Order("id").Find(&events).Error)
	require.Len(t, events, 2)
	require.Equal(t, "tool.started", events[0].EventType)
	require.Equal(t, call.CallID, events[0].CallID)
	require.Contains(t, events[0].Payload, `"arguments":{"query":"Go"}`)
	require.Equal(t, "tool.completed", events[1].EventType)
	require.Equal(t, call.CallID, events[1].CallID)
	require.Contains(t, events[1].Payload, `"resultRef"`)
}

// TestEinoMCPToolPersistsMessageCandidates 验证对应测试场景。
func TestEinoMCPToolPersistsMessageCandidates(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentRunTool{}, &model.ResourceSearch{}, &model.ResourceCandidate{}))
	client := &recordingMCPClient{textResult: true, result: map[string]any{
		"status": "success",
		"data": map[string]any{
			"dialog": map[string]any{"id": 7, "name": "Go 资料"},
			"messages": []any{map[string]any{"text": "preview", "file": map[string]any{
				"candidate_id": "signed-candidate", "name": "Go.pdf", "size": 123, "mime_type": "application/pdf",
			}}},
		},
	}}
	connection := &mcpConnection{Server: model.MCPServer{Code: "resources", RequestTimeoutSeconds: 15}, Client: client}
	definition := mcp.Tool{Name: "message_search", InputSchema: mcp.ToolInputSchema{Type: "object", Properties: map[string]any{
		"query": map[string]any{"type": "string"},
	}}}
	target, err := newTestMCPTool(db, connection, definition)
	require.NoError(t, err)
	target.scope = agenttool.Scope{OwnerID: 11}
	target.runID = 9

	_, err = target.InvokableRun(t.Context(), `{"query":"Go"}`)
	require.NoError(t, err)
	var search model.ResourceSearch
	require.NoError(t, db.First(&search).Error)
	require.EqualValues(t, 11, search.OwnerID)
	require.Equal(t, "Go", search.Query)
	var candidate model.ResourceCandidate
	require.NoError(t, db.First(&candidate).Error)
	require.Equal(t, search.ID, candidate.SearchID)
	require.Equal(t, "signed-candidate", candidate.ExternalID)
	require.Equal(t, "resources", candidate.MCPServerCode)
	require.JSONEq(t, `{"id":7,"name":"Go 资料"}`, candidate.Source)
	require.Equal(t, "preview", candidate.Preview)
	require.NotNil(t, candidate.ExpiresAt)
}

// newTestMCPTool 提供测试辅助能力。
func newTestMCPTool(db *gorm.DB, connection *mcpConnection, definition mcp.Tool) (*governedTool, error) {
	implementation, err := newMCPAgentTool(db, connection, definition)
	if err != nil {
		return nil, err
	}
	return &governedTool{db: db, implementation: implementation, harness: newAgentHarness(model.DefaultAgentConfig())}, nil
}

// recordingMCPClient 是测试辅助类型。
type recordingMCPClient struct {
	fakeMCPClient                // fakeMCPClient 提供默认模拟行为。
	calls         atomic.Int64   // calls 记录调用次数。
	lastName      string         // lastName 记录最近工具名。
	lastArguments map[string]any // lastArguments 记录最近调用参数。
	result        any            // result 是模拟结构化结果。
	textResult    bool           // textResult 控制返回文本结果。
}

// CallTool 提供测试辅助能力。
func (c *recordingMCPClient) CallTool(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	c.calls.Add(1)
	c.lastName = request.Params.Name
	c.lastArguments = request.GetArguments()
	result := c.result
	if result == nil {
		result = map[string]any{"ok": true}
	}
	if c.textResult {
		payload, _ := json.Marshal(result)
		return &mcp.CallToolResult{Content: []mcp.Content{mcp.NewTextContent(string(payload))}}, nil
	}
	return &mcp.CallToolResult{StructuredContent: result}, nil
}
