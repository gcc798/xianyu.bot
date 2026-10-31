package service

import (
	"context"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/xianyu/model"
	"github.com/glebarez/sqlite"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAgentRunnerApprovesResourceAcquisitionWithoutPersistingDeliveryToken(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.MCPServer{}, &model.AgentSession{}, &model.AgentRun{}, &model.AgentRunTool{}, &model.AgentEvent{}, &model.ResourceSearch{},
		&model.ResourceCandidate{}, &model.ResourceTask{},
	))
	oldDB := global.GVA_DB
	global.GVA_DB = db
	t.Cleanup(func() { global.GVA_DB = oldDB })
	server := model.MCPServer{Name: "资源服务", Code: "resources", BaseURL: "http://resources/mcp", Enabled: true, RequestTimeoutSeconds: 15, TaskTimeoutSeconds: 60}
	require.NoError(t, db.Create(&server).Error)
	search := model.ResourceSearch{OwnerID: 11, Query: "Go", Status: "completed"}
	require.NoError(t, db.Create(&search).Error)
	candidate := model.ResourceCandidate{SearchID: search.ID, MCPServerCode: server.Code, ExternalID: "signed-candidate", Name: "Go.pdf", Source: `{}`, License: "unknown"}
	require.NoError(t, db.Create(&candidate).Error)
	session := model.AgentSession{OwnerID: 11, Title: "审批轨迹"}
	require.NoError(t, db.Create(&session).Error)
	run := model.AgentRun{OwnerID: 11, SessionID: session.ID, ModelID: 1, ModelName: "test", Status: "completed", ConfigSnapshot: `{}`, StartedAt: time.Now()}
	require.NoError(t, db.Create(&run).Error)
	call := model.AgentRunTool{
		RunID: run.ID, CallID: "call-approve", MCPServerCode: server.Code, ToolName: "resource_acquire",
		Arguments: `{"candidate_id":"signed-candidate"}`, Status: "waiting_approval", StartedAt: time.Now(),
	}
	require.NoError(t, db.Create(&call).Error)
	client := &recordingMCPClient{textResult: true, result: map[string]any{
		"status": "success", "data": map[string]any{"task_id": "resource_acquire:42"},
	}}
	registry := &mcpRegistry{connect: func(context.Context, model.MCPServer) (*mcpConnection, error) {
		return &mcpConnection{Server: server, Client: client, Tools: []mcp.Tool{{Name: "resource_acquire"}}}, nil
	}}
	runner := &agentRunner{registry: registry, events: newAgentEventHub(db)}

	_, err = runner.Approve(t.Context(), 22, run.ID)
	require.Error(t, err)
	approved, err := runner.Approve(t.Context(), 11, run.ID)
	require.NoError(t, err)
	require.Equal(t, "completed", approved.Status)
	require.Equal(t, "resource_acquire", client.lastName)
	require.Equal(t, "signed-candidate", client.lastArguments["candidate_id"])
	require.NotEmpty(t, client.lastArguments["delivery_token"])
	require.NotZero(t, client.lastArguments["task_id"])

	var storedCall model.AgentRunTool
	require.NoError(t, db.First(&storedCall, call.ID).Error)
	require.Equal(t, `{"candidate_id":"signed-candidate"}`, storedCall.Arguments)
	require.NotContains(t, storedCall.Arguments, "delivery_token")
	require.Contains(t, storedCall.Result, "resource_acquire:42")
	var completedEvent model.AgentEvent
	require.NoError(t, db.Where("session_id = ? AND event_type = ?", session.ID, "tool.completed").First(&completedEvent).Error)
	require.Equal(t, call.CallID, completedEvent.CallID)
	var task model.ResourceTask
	require.NoError(t, db.First(&task).Error)
	require.EqualValues(t, 11, task.OwnerID)
	require.Equal(t, candidate.ID, task.CandidateID)
	require.Equal(t, "resource_acquire:42", task.ExternalTaskID)
	require.Equal(t, "running", task.State)
}
