package service

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model"
	"github.com/glebarez/sqlite"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestEinoMCPToolHidesDeliveryFieldsAndRequiresApproval(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentRunTool{}))
	oldDB := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = oldDB })
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
	target, err := newEinoMCPTool(connection, definition, newAgentHarness(model.DefaultAgentConfig()))
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

func TestEinoMCPToolCallsReadCapability(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentRunTool{}))
	oldDB := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = oldDB })
	client := new(recordingMCPClient)
	connection := &mcpConnection{
		Server: model.MCPServer{Code: "telegram", RequestTimeoutSeconds: 15}, Client: client,
	}
	definition := mcp.Tool{
		Name: "resource_search",
		InputSchema: mcp.ToolInputSchema{Type: "object", Properties: map[string]any{
			"query": map[string]any{"type": "string"},
		}},
	}
	target, err := newEinoMCPTool(connection, definition, newAgentHarness(model.DefaultAgentConfig()))
	require.NoError(t, err)
	target.runID = 9
	result, err := target.InvokableRun(t.Context(), `{"query":"Go"}`)
	require.NoError(t, err)
	require.Contains(t, result, "structuredContent")
	require.EqualValues(t, 1, client.calls.Load())
	require.Equal(t, "resource_search", client.lastName)
	var call model.AgentRunTool
	require.NoError(t, db.First(&call).Error)
	require.EqualValues(t, 9, call.RunID)
	require.Equal(t, `{"query":"Go"}`, call.Arguments)
	require.Equal(t, "completed", call.Status)
	require.Empty(t, call.ErrorMessage)
}

func TestEinoMCPToolPersistsResourceSearchCandidates(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentRunTool{}, &model.ResourceSearch{}, &model.ResourceCandidate{}))
	oldDB := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = oldDB })
	client := &recordingMCPClient{textResult: true, result: map[string]any{
		"status": "success",
		"data": map[string]any{"candidates": []any{map[string]any{
			"candidate_id": "signed-candidate", "name": "Go.pdf", "size": 123,
			"mime_type": "application/pdf", "source": map[string]any{"name": "Go 资料"},
			"preview": "preview", "license": "unknown", "risk": "许可未知",
		}}},
	}}
	connection := &mcpConnection{Server: model.MCPServer{Code: "resources", RequestTimeoutSeconds: 15}, Client: client}
	definition := mcp.Tool{Name: "resource_search", InputSchema: mcp.ToolInputSchema{Type: "object", Properties: map[string]any{
		"query": map[string]any{"type": "string"},
	}}}
	target, err := newEinoMCPTool(connection, definition, newAgentHarness(model.DefaultAgentConfig()))
	require.NoError(t, err)
	target.ownerID = 11
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
	require.NotNil(t, candidate.ExpiresAt)
}

type recordingMCPClient struct {
	fakeMCPClient
	calls         atomic.Int64
	lastName      string
	lastArguments map[string]any
	result        any
	textResult    bool
}

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
