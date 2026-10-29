package service

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xy/model"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func TestEinoMCPToolHidesDeliveryFieldsAndRequiresApproval(t *testing.T) {
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

	result, err := target.InvokableRun(t.Context(), `{"candidate_id":"candidate-1"}`)
	require.NoError(t, err)
	require.Contains(t, result, "approval_required")
	require.Zero(t, client.calls.Load())
}

func TestEinoMCPToolCallsReadCapability(t *testing.T) {
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
	result, err := target.InvokableRun(t.Context(), `{"query":"Go"}`)
	require.NoError(t, err)
	require.Contains(t, result, "structuredContent")
	require.EqualValues(t, 1, client.calls.Load())
	require.Equal(t, "resource_search", client.lastName)
}

type recordingMCPClient struct {
	fakeMCPClient
	calls    atomic.Int64
	lastName string
}

func (c *recordingMCPClient) CallTool(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	c.calls.Add(1)
	c.lastName = request.Params.Name
	return &mcp.CallToolResult{StructuredContent: map[string]any{"ok": true}}, nil
}
