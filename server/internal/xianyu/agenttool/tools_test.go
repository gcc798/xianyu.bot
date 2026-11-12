package xianyutool

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	agenttool "github.com/gcc798/xianyu.bot/internal/agent/tool"
	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/stretchr/testify/require"
)

// accountReaderStub 是测试辅助类型。
type accountReaderStub struct {
	ownerID uint // ownerID 记录调用用户。
}

type orderDetailStub struct {
	ownerID uint
	orderID uint
	err     error
}

type orderReadStub struct {
	orderReaderStub
	orderDetailStub
}

func (s *orderDetailStub) Get(_ context.Context, ownerID, orderID uint) (xyResponse.Order, error) {
	s.ownerID, s.orderID = ownerID, orderID
	return xyResponse.Order{ExternalOrderID: "order-1"}, s.err
}

type deliveryReadStub struct {
	ownerID  uint
	taskArgs xyRequest.DeliveryTaskSearch
	ruleID   uint
}

func (s *deliveryReadStub) ListTasks(_ context.Context, ownerID uint, input xyRequest.DeliveryTaskSearch) ([]xyResponse.DeliveryTask, int64, error) {
	s.ownerID, s.taskArgs = ownerID, input
	return []xyResponse.DeliveryTask{{ExternalOrderID: "order-1"}}, 1, nil
}

func (s *deliveryReadStub) Inventory(_ context.Context, ownerID, ruleID uint) (xyResponse.DeliveryInventory, error) {
	s.ownerID, s.ruleID = ownerID, ruleID
	return xyResponse.DeliveryInventory{RuleID: ruleID, Available: 3}, nil
}

func (*deliveryReadStub) Retry(context.Context, uint, uint) error { return nil }

type replyRuleReadStub struct {
	ownerID uint
	input   xyRequest.ReplyRuleSearch
}

func (s *replyRuleReadStub) List(_ context.Context, ownerID uint, input xyRequest.ReplyRuleSearch) ([]xyResponse.ReplyRule, int64, error) {
	s.ownerID, s.input = ownerID, input
	return []xyResponse.ReplyRule{{ProductTitle: "商品", Content: "回复"}}, 1, nil
}

func (*replyRuleReadStub) Save(context.Context, uint, xyRequest.SaveReplyRule) error { return nil }
func (*replyRuleReadStub) Delete(context.Context, uint, uint) error                  { return nil }

type deliveryRuleReadStub struct {
	ownerID uint
	input   xyRequest.DeliveryRuleSearch
	empty   bool
}

func (s *deliveryRuleReadStub) List(_ context.Context, ownerID uint, input xyRequest.DeliveryRuleSearch) ([]xyResponse.DeliveryRule, int64, error) {
	s.ownerID, s.input = ownerID, input
	if s.empty {
		return nil, 0, nil
	}
	return []xyResponse.DeliveryRule{{ID: 9, ProductTitle: "商品", Template: "已发货"}}, 1, nil
}

func (*deliveryRuleReadStub) Save(context.Context, uint, xyRequest.SaveDeliveryRule) error {
	return nil
}
func (*deliveryRuleReadStub) Delete(context.Context, uint, uint) error { return nil }

type businessEventReadStub struct {
	ownerID uint
	input   xyRequest.BusinessEventSearch
	err     error
}

func (s *businessEventReadStub) List(_ context.Context, ownerID uint, input xyRequest.BusinessEventSearch) ([]xyResponse.BusinessEvent, int64, error) {
	s.ownerID, s.input = ownerID, input
	return []xyResponse.BusinessEvent{{Summary: "发货失败"}}, 1, s.err
}

type relistReadStub struct {
	ownerID uint
	input   xyRequest.RelistTaskSearch
}

func (s *relistReadStub) ListTasks(_ context.Context, ownerID uint, input xyRequest.RelistTaskSearch) ([]model.RelistTask, int64, error) {
	s.ownerID, s.input = ownerID, input
	return []model.RelistTask{{Status: model.RelistTaskUnknown}}, 1, nil
}

func (*relistReadStub) Retry(context.Context, uint, uint) error { return nil }

// List 提供测试辅助能力。
func (s *accountReaderStub) List(_ context.Context, ownerID uint, _ xyRequest.AccountSearch) ([]model.Account, int64, error) {
	s.ownerID = ownerID
	return []model.Account{{Label: "主账号"}}, 1, nil
}

// productReaderStub 是测试辅助类型。
type productReaderStub struct{}

// List 提供测试辅助能力。
func (*productReaderStub) List(context.Context, uint, xyRequest.ProductSearch) ([]xyResponse.Product, int64, error) {
	return nil, 0, nil
}

func (*productReaderStub) Sync(context.Context, uint, xyRequest.ProductSync) (int, error) {
	return 0, nil
}
func (*productReaderStub) PatchSetting(context.Context, uint, xyRequest.PatchProductSetting) error {
	return nil
}

// orderReaderStub 是测试辅助类型。
type orderReaderStub struct{}

// List 提供测试辅助能力。
func (*orderReaderStub) List(context.Context, uint, xyRequest.OrderSearch) ([]xyResponse.Order, int64, error) {
	return nil, 0, nil
}

func (*orderReaderStub) SyncPending(context.Context, uint, uint) (int, error) { return 0, nil }

// messageServiceStub 是测试辅助类型。
type messageServiceStub struct {
	ownerID uint                  // ownerID 记录调用用户。
	reply   xyRequest.SendMessage // reply 记录回复请求。
}

// List 提供测试辅助能力。
func (*messageServiceStub) List(context.Context, uint, xyRequest.MessageSearch) ([]xyResponse.Message, int64, error) {
	return nil, 0, nil
}

// Reply 提供测试辅助能力。
func (s *messageServiceStub) Reply(_ context.Context, ownerID uint, input xyRequest.SendMessage) error {
	s.ownerID, s.reply = ownerID, input
	return nil
}

func (*messageServiceStub) UpdateTakeover(context.Context, uint, xyRequest.UpdateTakeover) error {
	return nil
}

// TestRegisterAddsStableXianyuToolDefinitions 验证对应测试场景。
func TestRegisterAddsStableXianyuToolDefinitions(t *testing.T) {
	registry := agenttool.NewRegistry()
	require.NoError(t, Register(registry, new(accountReaderStub), new(productReaderStub), new(orderReadStub), new(messageServiceStub), new(deliveryReadStub), new(replyRuleReadStub), new(deliveryRuleReadStub), new(businessEventReadStub), new(relistReadStub)))

	tools := registry.List()
	require.Len(t, tools, 23)
	require.Equal(t, "xianyu_account_list", tools[0].Definition().Code)
	accountSchema, err := json.Marshal(tools[0].Definition().Parameters)
	require.NoError(t, err)
	require.Contains(t, string(accountSchema), `"page"`)
	require.Contains(t, string(accountSchema), `"pageSize"`)
	reply, exists := registry.Get("xianyu_message_reply")
	require.True(t, exists)
	require.Equal(t, "回复闲鱼消息", reply.Definition().Name)
	require.True(t, reply.Definition().RequiresApproval)
	require.False(t, reply.Definition().ReadOnly)
	for _, code := range []string{"xianyu_product_sync", "xianyu_pending_order_sync", "xianyu_product_setting_patch", "xianyu_reply_rule_save", "xianyu_reply_rule_delete", "xianyu_delivery_rule_save", "xianyu_delivery_rule_delete", "xianyu_delivery_retry", "xianyu_relist_retry", "xianyu_message_takeover"} {
		tool, exists := registry.Get(code)
		require.True(t, exists, code)
		require.Equal(t, "local", tool.Definition().Source, code)
		require.True(t, tool.Definition().RequiresApproval, code)
		require.False(t, tool.Definition().ReadOnly, code)
	}
	for _, code := range []string{"xianyu_order_detail", "xianyu_delivery_task_list", "xianyu_delivery_inventory", "xianyu_reply_rule_list", "xianyu_delivery_rule_list", "xianyu_business_event_list", "xianyu_relist_task_list", "xianyu_delivery_import_handoff"} {
		tool, exists := registry.Get(code)
		require.True(t, exists, code)
		require.Equal(t, "local", tool.Definition().Source, code)
		require.True(t, tool.Definition().ReadOnly, code)
	}
	deliveryRule, exists := registry.Get("xianyu_delivery_rule_save")
	require.True(t, exists)
	schema, err := json.Marshal(deliveryRule.Definition().Parameters)
	require.NoError(t, err)
	require.NotContains(t, string(schema), "values")
	require.NotContains(t, string(schema), "cards")
}

// TestAccountToolUsesAuthenticatedOwner 验证对应测试场景。
func TestAccountToolUsesAuthenticatedOwner(t *testing.T) {
	accounts := new(accountReaderStub)
	registry := agenttool.NewRegistry()
	require.NoError(t, Register(registry, accounts, new(productReaderStub), new(orderReadStub), new(messageServiceStub), new(deliveryReadStub), new(replyRuleReadStub), new(deliveryRuleReadStub), new(businessEventReadStub), new(relistReadStub)))

	tool, exists := registry.Get("xianyu_account_list")
	require.True(t, exists)
	output, err := tool.Invoke(t.Context(), agenttool.Scope{OwnerID: 42}, []byte(`{"page":1,"pageSize":5}`))
	require.NoError(t, err)
	require.EqualValues(t, 42, accounts.ownerID)
	require.JSONEq(t, `{"list":[{"ID":0,"CreatedAt":"0001-01-01T00:00:00Z","UpdatedAt":"0001-01-01T00:00:00Z","externalUserId":"","label":"主账号","note":"","status":"","enabled":false,"hasCredential":false}],"total":1}`, output)
}

// TestMessageReplyToolDoesNotAcceptOwnerFromArguments 验证对应测试场景。
func TestMessageReplyToolDoesNotAcceptOwnerFromArguments(t *testing.T) {
	messages := new(messageServiceStub)
	tool := newMessageReply(messages)
	_, err := tool.Invoke(t.Context(), agenttool.Scope{OwnerID: 7}, []byte(`{"ownerId":99,"messageId":12,"text":" 已处理 "}`))
	require.NoError(t, err)
	require.EqualValues(t, 7, messages.ownerID)
	require.EqualValues(t, 12, messages.reply.MessageID)
	require.Equal(t, "已处理", messages.reply.Text)
}

func TestXianyuReadToolsUseOwnerScopeAndMapFilters(t *testing.T) {
	t.Run("order detail", func(t *testing.T) {
		service := new(orderDetailStub)
		tool := newOrderDetail(service)
		output, err := tool.Invoke(t.Context(), agenttool.Scope{OwnerID: 42}, []byte(`{"orderId":12}`))
		require.NoError(t, err)
		require.EqualValues(t, 42, service.ownerID)
		require.EqualValues(t, 12, service.orderID)
		require.JSONEq(t, `{"ID":0,"CreatedAt":"0001-01-01T00:00:00Z","UpdatedAt":"0001-01-01T00:00:00Z","accountId":0,"accountLabel":"","externalOrderId":"order-1","externalProductId":"","productTitle":"","buyerUserId":"","buyerUserName":"","sessionId":"","skuKey":"","status":"","amountCent":0}`, output)
	})
	t.Run("delivery tasks and inventory", func(t *testing.T) {
		service := new(deliveryReadStub)
		listTool := newDeliveryTaskList(service)
		_, err := listTool.Invoke(t.Context(), agenttool.Scope{OwnerID: 42}, []byte(`{"page":2,"pageSize":5,"accountId":3,"productId":4,"status":"failed","keyword":"order"}`))
		require.NoError(t, err)
		require.EqualValues(t, 42, service.ownerID)
		require.Equal(t, xyRequest.DeliveryTaskSearch{
			PageInfo:  xyRequest.PageInfo{Pagination: xyRequest.Pagination{Page: 2, PageSize: 5}},
			AccountID: 3, ProductID: 4, Status: "failed", Keyword: "order",
		}, service.taskArgs)
		inventoryTool := newDeliveryInventory(service)
		output, err := inventoryTool.Invoke(t.Context(), agenttool.Scope{OwnerID: 42}, []byte(`{"ruleId":9}`))
		require.NoError(t, err)
		require.EqualValues(t, 42, service.ownerID)
		require.EqualValues(t, 9, service.ruleID)
		require.JSONEq(t, `{"ruleId":9,"available":3,"reserved":0,"consumed":0}`, output)
	})
	t.Run("rule filters", func(t *testing.T) {
		replyRules := new(replyRuleReadStub)
		_, err := newReplyRuleList(replyRules).Invoke(t.Context(), agenttool.Scope{OwnerID: 42}, []byte(`{"page":3,"pageSize":7,"productId":8,"type":"fuzzy","keyword":"size","enabledOnly":true}`))
		require.NoError(t, err)
		require.EqualValues(t, 42, replyRules.ownerID)
		require.Equal(t, xyRequest.ReplyRuleSearch{
			PageInfo:  xyRequest.PageInfo{Pagination: xyRequest.Pagination{Page: 3, PageSize: 7}, Keyword: "size"},
			ProductID: 8, Type: "fuzzy", EnabledOnly: true,
		}, replyRules.input)

		deliveryRules := new(deliveryRuleReadStub)
		_, err = newDeliveryRuleList(deliveryRules).Invoke(t.Context(), agenttool.Scope{OwnerID: 42}, []byte(`{"page":1,"pageSize":10,"ruleId":9,"productId":8,"mode":"card","keyword":"SKU","enabledOnly":true}`))
		require.NoError(t, err)
		require.EqualValues(t, 42, deliveryRules.ownerID)
		require.Equal(t, xyRequest.DeliveryRuleSearch{
			PageInfo: xyRequest.PageInfo{Pagination: xyRequest.Pagination{Page: 1, PageSize: 10}, Keyword: "SKU"},
			RuleID:   9, ProductID: 8, Mode: "card", EnabledOnly: true,
		}, deliveryRules.input)
		output, err := newDeliveryImportHandoff(deliveryRules).Invoke(t.Context(), agenttool.Scope{OwnerID: 42}, []byte(`{"ruleId":9}`))
		require.NoError(t, err)
		require.JSONEq(t, `{"path":"/xianyu/delivery-rules?importRuleId=9"}`, output)
		require.EqualValues(t, 42, deliveryRules.ownerID)
		require.EqualValues(t, 9, deliveryRules.input.RuleID)
		require.NotContains(t, output, "card")
		deliveryRules.empty = true
		_, err = newDeliveryImportHandoff(deliveryRules).Invoke(t.Context(), agenttool.Scope{OwnerID: 42}, []byte(`{"ruleId":99}`))
		require.ErrorContains(t, err, "未找到当前用户拥有的发货规则")
	})
	t.Run("business events and relist tasks", func(t *testing.T) {
		events := new(businessEventReadStub)
		_, err := newBusinessEventList(events).Invoke(t.Context(), agenttool.Scope{OwnerID: 42}, []byte(`{"page":4,"pageSize":6,"accountId":2,"productId":3,"orderId":4,"type":"delivery","status":"failed","keyword":"network"}`))
		require.NoError(t, err)
		require.EqualValues(t, 42, events.ownerID)
		require.Equal(t, xyRequest.BusinessEventSearch{
			PageInfo:  xyRequest.PageInfo{Pagination: xyRequest.Pagination{Page: 4, PageSize: 6}, Keyword: "network"},
			AccountID: 2, ProductID: 3, OrderID: 4, Type: "delivery", Status: "failed",
		}, events.input)

		relist := new(relistReadStub)
		_, err = newRelistTaskList(relist).Invoke(t.Context(), agenttool.Scope{OwnerID: 42}, []byte(`{"page":2,"pageSize":8,"accountId":5,"productId":6,"status":"unknown"}`))
		require.NoError(t, err)
		require.EqualValues(t, 42, relist.ownerID)
		require.Equal(t, xyRequest.RelistTaskSearch{
			Pagination: xyRequest.Pagination{Page: 2, PageSize: 8}, AccountID: 5, ProductID: 6, Status: "unknown",
		}, relist.input)
	})
}

func TestReadToolDefinitionsAreReadOnlyAndDoNotExposeOwnerArgument(t *testing.T) {
	definitions := []agenttool.Tool{
		newOrderDetail(new(orderDetailStub)),
		newDeliveryTaskList(new(deliveryReadStub)), newDeliveryInventory(new(deliveryReadStub)),
		newReplyRuleList(new(replyRuleReadStub)), newDeliveryRuleList(new(deliveryRuleReadStub)),
		newBusinessEventList(new(businessEventReadStub)), newRelistTaskList(new(relistReadStub)),
	}
	for _, implementation := range definitions {
		definition := implementation.Definition()
		require.True(t, definition.ReadOnly, definition.Code)
		require.False(t, definition.RequiresApproval, definition.Code)
		schema, err := json.Marshal(definition.Parameters)
		require.NoError(t, err)
		require.NotContains(t, string(schema), "ownerId", definition.Code)
	}

	events := &businessEventReadStub{err: errors.New("private service detail")}
	_, err := newBusinessEventList(events).Invoke(t.Context(), agenttool.Scope{OwnerID: 42}, []byte(`{}`))
	executionErr, ok := errors.AsType[*agenttool.ExecutionError](err)
	require.True(t, ok)
	require.Equal(t, "xianyu_operation_failed", executionErr.Code)
}
