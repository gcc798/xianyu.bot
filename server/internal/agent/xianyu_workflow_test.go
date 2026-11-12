package agent

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	agenttool "github.com/gcc798/xianyu.bot/internal/agent/tool"
	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/response"
	xianyutool "github.com/gcc798/xianyu.bot/internal/xianyu/agenttool"
	"github.com/stretchr/testify/require"
)

type workflowAccountStub struct{}

func (workflowAccountStub) List(context.Context, uint, xyRequest.AccountSearch) ([]model.Account, int64, error) {
	return nil, 0, nil
}

type workflowProductStub struct {
	ownerID uint
	patch   xyRequest.PatchProductSetting
	calls   int
}

func (*workflowProductStub) List(context.Context, uint, xyRequest.ProductSearch) ([]xyResponse.Product, int64, error) {
	return nil, 0, nil
}

func (*workflowProductStub) Sync(context.Context, uint, xyRequest.ProductSync) (int, error) {
	return 0, nil
}

func (s *workflowProductStub) PatchSetting(_ context.Context, ownerID uint, patch xyRequest.PatchProductSetting) error {
	s.ownerID, s.patch = ownerID, patch
	s.calls++
	return nil
}

type workflowOrderStub struct{}

func (workflowOrderStub) List(context.Context, uint, xyRequest.OrderSearch) ([]xyResponse.Order, int64, error) {
	return nil, 0, nil
}

func (workflowOrderStub) Get(context.Context, uint, uint) (xyResponse.Order, error) {
	return xyResponse.Order{}, nil
}

func (workflowOrderStub) SyncPending(context.Context, uint, uint) (int, error) { return 0, nil }

type workflowMessageStub struct{ takeoverCalls int }

func (*workflowMessageStub) List(context.Context, uint, xyRequest.MessageSearch) ([]xyResponse.Message, int64, error) {
	return nil, 0, nil
}

func (*workflowMessageStub) Reply(context.Context, uint, xyRequest.SendMessage) error { return nil }

func (s *workflowMessageStub) UpdateTakeover(context.Context, uint, xyRequest.UpdateTakeover) error {
	s.takeoverCalls++
	return nil
}

type workflowDeliveryStub struct {
	ownerID uint
	taskID  uint
	calls   int
	events  *workflowEventStub
}

func (s *workflowDeliveryStub) ListTasks(_ context.Context, ownerID uint, _ xyRequest.DeliveryTaskSearch) ([]xyResponse.DeliveryTask, int64, error) {
	if s.calls == 0 {
		return []xyResponse.DeliveryTask{{ID: 83, Status: "failed", LastError: "首次发货失败"}}, 1, nil
	}
	return []xyResponse.DeliveryTask{{ID: 83, Status: "pending", Attempts: 2}}, 1, nil
}

func (*workflowDeliveryStub) Inventory(context.Context, uint, uint) (xyResponse.DeliveryInventory, error) {
	return xyResponse.DeliveryInventory{}, nil
}

func (s *workflowDeliveryStub) Retry(_ context.Context, ownerID, taskID uint) error {
	s.ownerID, s.taskID = ownerID, taskID
	s.calls++
	s.events.events = append(s.events.events, xyResponse.BusinessEvent{Type: "delivery", Status: "success", Summary: "发货任务已重试"})
	return nil
}

type workflowReplyRuleStub struct {
	ownerID uint
	input   xyRequest.SaveReplyRule
	calls   int
}

func (*workflowReplyRuleStub) List(context.Context, uint, xyRequest.ReplyRuleSearch) ([]xyResponse.ReplyRule, int64, error) {
	return nil, 0, nil
}

func (s *workflowReplyRuleStub) Save(_ context.Context, ownerID uint, input xyRequest.SaveReplyRule) error {
	s.ownerID, s.input = ownerID, input
	s.calls++
	return nil
}
func (*workflowReplyRuleStub) Delete(context.Context, uint, uint) error { return nil }

type workflowDeliveryRuleStub struct{}

func (*workflowDeliveryRuleStub) List(context.Context, uint, xyRequest.DeliveryRuleSearch) ([]xyResponse.DeliveryRule, int64, error) {
	return nil, 0, nil
}

func (*workflowDeliveryRuleStub) Save(context.Context, uint, xyRequest.SaveDeliveryRule) error {
	return nil
}
func (*workflowDeliveryRuleStub) Delete(context.Context, uint, uint) error { return nil }

type workflowEventStub struct {
	ownerID uint
	calls   int
	events  []xyResponse.BusinessEvent
}

func (s *workflowEventStub) List(_ context.Context, ownerID uint, _ xyRequest.BusinessEventSearch) ([]xyResponse.BusinessEvent, int64, error) {
	s.ownerID = ownerID
	s.calls++
	return s.events, int64(len(s.events)), nil
}

type workflowRelistStub struct{}

func (*workflowRelistStub) ListTasks(context.Context, uint, xyRequest.RelistTaskSearch) ([]model.RelistTask, int64, error) {
	return nil, 0, nil
}

func (*workflowRelistStub) Retry(context.Context, uint, uint) error { return nil }

type controlledXianyuModel struct {
	toolCall schema.ToolCall
	answer   string
	calls    atomic.Int64
}

func (*controlledXianyuModel) Generate(context.Context, []*schema.Message, ...einomodel.Option) (*schema.Message, error) {
	return nil, errors.New("unexpected non-streaming call")
}

func (m *controlledXianyuModel) Stream(context.Context, []*schema.Message, ...einomodel.Option) (*schema.StreamReader[*schema.Message], error) {
	if m.calls.Add(1) == 1 {
		return schema.StreamReaderFromArray([]*schema.Message{schema.AssistantMessage("", []schema.ToolCall{m.toolCall})}), nil
	}
	return schema.StreamReaderFromArray([]*schema.Message{schema.AssistantMessage(m.answer, nil)}), nil
}

func (m *controlledXianyuModel) reset(call schema.ToolCall, answer string) {
	m.toolCall, m.answer = call, answer
	m.calls.Store(0)
}

func TestRunnerApprovesAndRejectsXianyuWriteTools(t *testing.T) {
	db := agentRunnerTestDB(t)
	product := new(workflowProductStub)
	replyRules := new(workflowReplyRuleStub)
	events := &workflowEventStub{events: []xyResponse.BusinessEvent{{Type: "automation", Status: "success", Summary: "自动化运行正常"}}}
	delivery := &workflowDeliveryStub{events: events}
	messages := new(workflowMessageStub)
	registry := agenttool.NewRegistry()
	require.NoError(t, xianyutool.Register(
		registry, workflowAccountStub{}, product, workflowOrderStub{}, messages, delivery,
		replyRules, new(workflowDeliveryRuleStub), events, new(workflowRelistStub),
	))
	chatModel := &controlledXianyuModel{answer: "等待确认"}
	runner := &Runner{
		db: db, registry: &MCPRegistry{db: db}, tools: registry, events: NewEventHub(db),
		newModel: func(context.Context, model.AIModel, model.AgentConfig) (einomodel.BaseChatModel, error) {
			return chatModel, nil
		},
	}

	inspect := schema.ToolCall{ID: "inspect-events", Type: "function", Function: schema.FunctionCall{
		Name: "xianyu_business_event_list", Arguments: `{"page":1,"pageSize":10}`,
	}}
	inspectionSession := model.AgentSession{OwnerID: 11, Title: "查看自动化状态"}
	require.NoError(t, db.Create(&inspectionSession).Error)
	chatModel.reset(inspect, "自动化运行正常")
	inspectionRun, err := runner.Run(t.Context(), 11, inspectionSession.ID, "查看最近闲鱼自动化状态", "inspect-run")
	require.NoError(t, err)
	require.Equal(t, "自动化运行正常", inspectionRun.Assistant.Content)
	require.EqualValues(t, 11, events.ownerID)
	require.Equal(t, 1, events.calls)
	var inspectionTool model.AgentRunTool
	require.NoError(t, db.Where("run_id = ?", inspectionRun.Run.ID).First(&inspectionTool).Error)
	require.Equal(t, "completed", inspectionTool.Status)
	var inspected struct {
		List []xyResponse.BusinessEvent `json:"list"`
	}
	require.NoError(t, json.Unmarshal([]byte(inspectionTool.Result), &inspected))
	require.Equal(t, "自动化运行正常", inspected.List[0].Summary)

	patch := schema.ToolCall{ID: "patch-setting", Type: "function", Function: schema.FunctionCall{
		Name: "xianyu_product_setting_patch", Arguments: `{"productId":71,"autoReplyEnabled":false}`,
	}}
	session := model.AgentSession{OwnerID: 11, Title: "修改自动化设置"}
	require.NoError(t, db.Create(&session).Error)
	chatModel.reset(patch, "已提交自动化设置变更审批")
	patchRun, err := runner.Run(t.Context(), 11, session.ID, "关闭商品自动回复", "patch-run")
	require.NoError(t, err)
	require.Zero(t, product.calls)
	var patchCall model.AgentRunTool
	require.NoError(t, db.Where("run_id = ?", patchRun.Run.ID).First(&patchCall).Error)
	require.Equal(t, "waiting_approval", patchCall.Status)
	approvedPatch, err := runner.Approve(t.Context(), 11, patchRun.Run.ID)
	require.NoError(t, err)
	require.Equal(t, "completed", approvedPatch.Status)
	require.Equal(t, 1, product.calls)
	require.EqualValues(t, 11, product.ownerID)
	require.EqualValues(t, 71, product.patch.ProductID)
	require.NotNil(t, product.patch.AutoReplyEnabled)
	require.False(t, *product.patch.AutoReplyEnabled)
	require.Nil(t, product.patch.AutoDeliveryEnabled)
	require.Nil(t, product.patch.AutoRelistEnabled)

	ruleSave := schema.ToolCall{ID: "save-reply-rule", Type: "function", Function: schema.FunctionCall{
		Name: "xianyu_reply_rule_save", Arguments: `{"productId":71,"type":"fuzzy","keyword":"物流","priority":2,"enabled":true,"content":"我会尽快为你处理"}`,
	}}
	ruleSession := model.AgentSession{OwnerID: 11, Title: "创建自动回复规则"}
	require.NoError(t, db.Create(&ruleSession).Error)
	chatModel.reset(ruleSave, "创建自动回复规则等待确认")
	ruleRun, err := runner.Run(t.Context(), 11, ruleSession.ID, "创建物流关键词自动回复规则", "rule-run")
	require.NoError(t, err)
	require.Zero(t, replyRules.calls)
	approvedRule, err := runner.Approve(t.Context(), 11, ruleRun.Run.ID)
	require.NoError(t, err)
	require.Equal(t, "completed", approvedRule.Status)
	require.Equal(t, 1, replyRules.calls)
	require.EqualValues(t, 11, replyRules.ownerID)
	require.Equal(t, "物流", replyRules.input.Keyword)
	require.Equal(t, "我会尽快为你处理", replyRules.input.Content)
	require.Contains(t, approvedRule.Result, `"ok":true`)

	retry := schema.ToolCall{ID: "retry-delivery", Type: "function", Function: schema.FunctionCall{
		Name: "xianyu_delivery_retry", Arguments: `{"taskId":83}`,
	}}
	secondSession := model.AgentSession{OwnerID: 11, Title: "重试发货任务"}
	require.NoError(t, db.Create(&secondSession).Error)
	chatModel.reset(retry, "发货重试等待确认")
	retryRun, err := runner.Run(t.Context(), 11, secondSession.ID, "重试发货任务 83", "retry-run")
	require.NoError(t, err)
	require.Zero(t, delivery.calls)
	approvedRetry, err := runner.Approve(t.Context(), 11, retryRun.Run.ID)
	require.NoError(t, err)
	require.Equal(t, "completed", approvedRetry.Status)
	require.Equal(t, 1, delivery.calls)
	require.EqualValues(t, 11, delivery.ownerID)
	require.EqualValues(t, 83, delivery.taskID)
	require.Contains(t, approvedRetry.Result, `"ok":true`)
	var retryEventCount int64
	require.NoError(t, db.Model(&model.AgentEvent{}).Where("session_id = ? AND event_type = ?", secondSession.ID, "tool.completed").Count(&retryEventCount).Error)
	require.EqualValues(t, 1, retryEventCount)
	var completedEvent model.AgentEvent
	require.NoError(t, db.Where("session_id = ? AND event_type = ?", secondSession.ID, "tool.completed").First(&completedEvent).Error)
	require.Contains(t, completedEvent.Payload, "xianyu_delivery_retry")

	verify := schema.ToolCall{ID: "verify-delivery", Type: "function", Function: schema.FunctionCall{
		Name: "xianyu_delivery_task_list", Arguments: `{"page":1,"pageSize":10,"status":"pending"}`,
	}}
	verifySession := model.AgentSession{OwnerID: 11, Title: "核对发货任务结果"}
	require.NoError(t, db.Create(&verifySession).Error)
	chatModel.reset(verify, "发货任务已进入待处理状态")
	verifyRun, err := runner.Run(t.Context(), 11, verifySession.ID, "核对发货任务状态", "verify-run")
	require.NoError(t, err)
	var verifyTool model.AgentRunTool
	require.NoError(t, db.Where("run_id = ?", verifyRun.Run.ID).First(&verifyTool).Error)
	var verified struct {
		List []xyResponse.DeliveryTask `json:"list"`
	}
	require.NoError(t, json.Unmarshal([]byte(verifyTool.Result), &verified))
	require.Equal(t, "pending", verified.List[0].Status)
	verifyEvent := inspect
	verifyEvent.ID = "verify-events"
	chatModel.reset(verifyEvent, "重试事件已记录")
	verifiedEvents, err := runner.Run(t.Context(), 11, inspectionSession.ID, "确认刚才的发货事件", "verify-event-run")
	require.NoError(t, err)
	require.Equal(t, "重试事件已记录", verifiedEvents.Assistant.Content)
	require.Equal(t, 2, events.calls)
	require.Equal(t, "发货任务已重试", events.events[len(events.events)-1].Summary)

	takeover := schema.ToolCall{ID: "reject-takeover", Type: "function", Function: schema.FunctionCall{
		Name: "xianyu_message_takeover", Arguments: `{"messageId":92,"active":true}`,
	}}
	thirdSession := model.AgentSession{OwnerID: 11, Title: "拒绝接管变更"}
	require.NoError(t, db.Create(&thirdSession).Error)
	chatModel.reset(takeover, "等待确认")
	rejectedRun, err := runner.Run(t.Context(), 11, thirdSession.ID, "开启人工接管", "reject-run")
	require.NoError(t, err)
	rejected, err := runner.Reject(t.Context(), 11, rejectedRun.Run.ID)
	require.NoError(t, err)
	require.Equal(t, "rejected", rejected.Status)
	require.Zero(t, messages.takeoverCalls)
	var rejectionCount int64
	require.NoError(t, db.Model(&model.AgentEvent{}).Where("session_id = ? AND event_type = ?", thirdSession.ID, "approval.rejected").Count(&rejectionCount).Error)
	require.EqualValues(t, 1, rejectionCount)
}
