package agent

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/eino-contrib/jsonschema"
	agenttool "github.com/gcc798/xianyu.bot/internal/agent/tool"
	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/glebarez/sqlite"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// approvalToolStub 是测试辅助类型。
type approvalToolStub struct {
	ownerID uint // ownerID 记录工具调用用户。
	calls   int  // calls 记录业务调用次数。
}

// Definition 提供测试辅助能力。
func (*approvalToolStub) Definition() agenttool.Definition {
	return agenttool.Definition{
		Code: "xianyu_message_reply", Name: "回复闲鱼消息", Source: "local",
		Description: "测试回复工具", Parameters: &jsonschema.Schema{Type: "object"},
		RequiresApproval: true, Timeout: time.Second,
	}
}

// Invoke 提供测试辅助能力。
func (s *approvalToolStub) Invoke(_ context.Context, scope agenttool.Scope, _ json.RawMessage) (string, error) {
	s.ownerID = scope.OwnerID
	s.calls++
	return `{"ok":true}`, nil
}

// TestAgentRunnerApprovesResourceAcquisitionWithoutPersistingDeliveryToken 验证对应测试场景。
func TestAgentRunnerApprovesResourceAcquisitionWithoutPersistingDeliveryToken(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.MCPServer{}, &model.AgentSession{}, &model.AgentRun{}, &model.AgentRunTool{}, &model.AgentEvent{}, &model.ResourceSearch{},
		&model.ResourceCandidate{}, &model.ResourceTask{},
	))
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
	registry := &MCPRegistry{db: db, connect: func(context.Context, model.MCPServer) (*mcpConnection, error) {
		return &mcpConnection{Server: server, Client: client, Tools: []mcp.Tool{{Name: "resource_acquire"}}}, nil
	}}
	resources := &ResourceCatalog{db: db, events: NewEventHub(db), registry: registry}
	runner := &Runner{db: db, registry: registry, resources: resources, events: resources.events}

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

// TestAgentRunnerApprovesLocalToolWithRunOwner 验证对应测试场景。
func TestAgentRunnerApprovesLocalToolWithRunOwner(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentSession{}, &model.AgentRun{}, &model.AgentRunTool{}, &model.AgentEvent{}))
	session := model.AgentSession{OwnerID: 11, Title: "本地工具审批"}
	require.NoError(t, db.Create(&session).Error)
	run := model.AgentRun{OwnerID: 11, SessionID: session.ID, ModelID: 1, ModelName: "test", Status: "completed", ConfigSnapshot: `{}`, StartedAt: time.Now()}
	require.NoError(t, db.Create(&run).Error)
	call := model.AgentRunTool{
		RunID: run.ID, CallID: "local-approve", ToolCode: "xianyu_message_reply", ToolName: "回复闲鱼消息", ToolSource: "local",
		Arguments: `{"messageId":12,"text":"已处理"}`, Status: "waiting_approval", StartedAt: time.Now(),
	}
	require.NoError(t, db.Create(&call).Error)
	implementation := new(approvalToolStub)
	registry := agenttool.NewRegistry()
	require.NoError(t, registry.Register(implementation))
	runner := &Runner{db: db, tools: registry, events: NewEventHub(db)}

	approved, err := runner.Approve(t.Context(), 11, run.ID)
	require.NoError(t, err)
	require.Equal(t, "completed", approved.Status)
	require.EqualValues(t, 11, implementation.ownerID)
	var stored model.AgentRunTool
	require.NoError(t, db.First(&stored, call.ID).Error)
	require.JSONEq(t, `{"ok":true}`, stored.Result)
}

func TestAgentRunnerRejectsLocalToolWithoutExecution(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentSession{}, &model.AgentRun{}, &model.AgentRunTool{}, &model.AgentEvent{}))
	session := model.AgentSession{OwnerID: 11, Title: "拒绝审批"}
	require.NoError(t, db.Create(&session).Error)
	run := model.AgentRun{OwnerID: 11, SessionID: session.ID, ModelID: 1, ModelName: "test", Status: "completed", ConfigSnapshot: `{}`, StartedAt: time.Now()}
	require.NoError(t, db.Create(&run).Error)
	call := model.AgentRunTool{
		RunID: run.ID, CallID: "local-reject", ToolCode: "xianyu_message_reply", ToolName: "回复闲鱼消息", ToolSource: "local",
		Arguments: `{"messageId":12,"text":"已处理"}`, Status: "waiting_approval", StartedAt: time.Now(),
	}
	require.NoError(t, db.Create(&call).Error)
	implementation := new(approvalToolStub)
	registry := agenttool.NewRegistry()
	require.NoError(t, registry.Register(implementation))
	events := NewEventHub(db)
	runner := &Runner{db: db, tools: registry, events: events}

	_, err = runner.Reject(t.Context(), 22, run.ID)
	require.Error(t, err)
	rejected, err := runner.Reject(t.Context(), 11, run.ID)
	require.NoError(t, err)
	require.Equal(t, "rejected", rejected.Status)
	require.Zero(t, implementation.calls)
	var stored model.AgentRunTool
	require.NoError(t, db.First(&stored, call.ID).Error)
	require.Equal(t, "approval_rejected", stored.ErrorCode)
	var audit model.AgentEvent
	require.NoError(t, db.Where("session_id = ? AND event_type = ?", session.ID, "approval.rejected").First(&audit).Error)
	require.Contains(t, audit.Payload, call.CallID)
}

func TestGovernedToolQueuesWriteBeforeCallingService(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AgentRunTool{}))
	implementation := new(approvalToolStub)
	target := &governedTool{
		db: db, implementation: implementation, harness: newAgentHarness(model.DefaultAgentConfig()),
		scope: agenttool.Scope{OwnerID: 11}, runID: 9,
	}
	result, err := target.InvokableRun(t.Context(), `{"messageId":12,"text":"已处理"}`)
	require.NoError(t, err)
	require.Contains(t, result, "approval_required")
	require.Zero(t, implementation.calls)
	require.Equal(t, 1, target.harness.Usage().Writes)
	var call model.AgentRunTool
	require.NoError(t, db.Where("run_id = ?", 9).First(&call).Error)
	require.Equal(t, "waiting_approval", call.Status)
	require.Equal(t, `{"messageId":12,"text":"已处理"}`, call.Arguments)
}
