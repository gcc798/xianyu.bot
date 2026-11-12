package xianyutool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/eino-contrib/jsonschema"
	agenttool "github.com/gcc798/xianyu.bot/internal/agent/tool"
	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/response"
)

// accountReader 描述账号查询工具需要的最小服务能力。
type accountReader interface {
	// List 分页查询指定用户的闲鱼账号。
	List(context.Context, uint, xyRequest.AccountSearch) ([]model.Account, int64, error)
}

// productReader 描述商品查询工具需要的最小服务能力。
type productReader interface {
	// List 分页查询指定用户的闲鱼商品。
	List(context.Context, uint, xyRequest.ProductSearch) ([]xyResponse.Product, int64, error)
}

// orderReader 描述订单查询工具需要的最小服务能力。
type orderReader interface {
	// List 分页查询指定用户的闲鱼订单。
	List(context.Context, uint, xyRequest.OrderSearch) ([]xyResponse.Order, int64, error)
}

type orderDetailReader interface {
	Get(context.Context, uint, uint) (xyResponse.Order, error)
}

type deliveryReader interface {
	ListTasks(context.Context, uint, xyRequest.DeliveryTaskSearch) ([]xyResponse.DeliveryTask, int64, error)
	Inventory(context.Context, uint, uint) (xyResponse.DeliveryInventory, error)
}

type replyRuleReader interface {
	List(context.Context, uint, xyRequest.ReplyRuleSearch) ([]xyResponse.ReplyRule, int64, error)
}

type deliveryRuleReader interface {
	List(context.Context, uint, xyRequest.DeliveryRuleSearch) ([]xyResponse.DeliveryRule, int64, error)
}

type businessEventReader interface {
	List(context.Context, uint, xyRequest.BusinessEventSearch) ([]xyResponse.BusinessEvent, int64, error)
}

type relistReader interface {
	ListTasks(context.Context, uint, xyRequest.RelistTaskSearch) ([]model.RelistTask, int64, error)
}

type productWriter interface {
	productReader
	Sync(context.Context, uint, xyRequest.ProductSync) (int, error)
	PatchSetting(context.Context, uint, xyRequest.PatchProductSetting) error
}

type orderWriter interface {
	SyncPending(context.Context, uint, uint) (int, error)
}

type orderService interface {
	orderReader
	orderDetailReader
	orderWriter
}

type replyRuleWriter interface {
	Save(context.Context, uint, xyRequest.SaveReplyRule) error
	Delete(context.Context, uint, uint) error
}

type replyRuleService interface {
	replyRuleReader
	replyRuleWriter
}

type deliveryRuleWriter interface {
	Save(context.Context, uint, xyRequest.SaveDeliveryRule) error
	Delete(context.Context, uint, uint) error
}

type deliveryRuleService interface {
	deliveryRuleReader
	deliveryRuleWriter
}

type deliveryWriter interface {
	Retry(context.Context, uint, uint) error
}

type deliveryService interface {
	deliveryReader
	deliveryWriter
}

type relistWriter interface {
	Retry(context.Context, uint, uint) error
}

type relistService interface {
	relistReader
	relistWriter
}

// messageService 描述消息查询与回复工具需要的服务能力。
type messageService interface {
	// List 分页查询指定用户的闲鱼消息。
	List(context.Context, uint, xyRequest.MessageSearch) ([]xyResponse.Message, int64, error)
	// Reply 回复指定用户拥有的入站消息。
	Reply(context.Context, uint, xyRequest.SendMessage) error
	UpdateTakeover(context.Context, uint, xyRequest.UpdateTakeover) error
}

// xianyuTool 将闲鱼业务函数实现为统一 Agent 工具。
type xianyuTool struct {
	definition agenttool.Definition                                                    // definition 描述工具身份、参数和权限。
	invoke     func(context.Context, agenttool.Scope, json.RawMessage) (string, error) // invoke 执行业务能力。
}

// Definition 返回工具不可变定义。
func (t *xianyuTool) Definition() agenttool.Definition { return t.definition }

// Invoke 在 scope 用户边界内执行工具。
func (t *xianyuTool) Invoke(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
	return t.invoke(ctx, scope, arguments)
}

// Register 将全部闲鱼本地工具注册到 registry，并拒绝缺失依赖。
func Register(registry *agenttool.Registry, accounts accountReader, products productWriter, orders orderService, messages messageService, delivery deliveryService, replyRules replyRuleService, deliveryRules deliveryRuleService, events businessEventReader, relists relistService) error {
	if registry == nil || accounts == nil || products == nil || orders == nil || messages == nil || delivery == nil || replyRules == nil || deliveryRules == nil || events == nil || relists == nil {
		return errors.New("闲鱼工具依赖不能为空")
	}
	tools := []*xianyuTool{
		newAccountList(accounts), newProductList(products), newOrderList(orders),
		newMessageList(messages), newMessageReply(messages),
		newOrderDetail(orders), newDeliveryTaskList(delivery), newDeliveryInventory(delivery),
		newReplyRuleList(replyRules), newDeliveryRuleList(deliveryRules), newBusinessEventList(events), newRelistTaskList(relists), newDeliveryImportHandoff(deliveryRules),
		newProductSync(products), newOrderSync(orders), newProductSettingPatch(products),
		newReplyRuleSave(replyRules), newReplyRuleDelete(replyRules), newDeliveryRuleSave(deliveryRules), newDeliveryRuleDelete(deliveryRules),
		newDeliveryRetry(delivery), newRelistRetry(relists), newMessageTakeover(messages),
	}
	for _, implementation := range tools {
		if err := registry.Register(implementation); err != nil {
			return err
		}
	}
	return nil
}

// pagination 是查询工具共享的分页输入。
type pagination struct {
	Page     int `json:"page" jsonschema_description:"页码，默认 1"`               // Page 是从 1 开始的页码。
	PageSize int `json:"pageSize" jsonschema_description:"每页数量，最大 100，默认 10"` // PageSize 是每页数量。
}

// accountListInput 是账号查询工具输入。
type accountListInput struct {
	pagination        // pagination 提供页码与每页数量。
	Keyword    string `json:"keyword" jsonschema_description:"账号标签、备注或闲鱼用户 ID 关键词"` // Keyword 是账号模糊检索词。
	Status     string `json:"status" jsonschema_description:"账号状态"`                 // Status 按账号状态筛选。
	Enabled    *bool  `json:"enabled,omitempty" jsonschema_description:"是否启用"`      // Enabled 按启用状态筛选。
}

// newAccountList 创建只读账号查询工具。
func newAccountList(accounts accountReader) *xianyuTool {
	return &xianyuTool{definition: readDefinition(
		"xianyu_account_list", "查询闲鱼账号", "查询当前用户的闲鱼账号档案，不返回 Cookie、Token 等凭据。", accountListInput{},
	), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input accountListInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		accounts, total, err := accounts.List(ctx, scope.OwnerID, xyRequest.AccountSearch{
			PageInfo: xyRequest.PageInfo{Pagination: page(input.pagination), Keyword: input.Keyword},
			Status:   input.Status, Enabled: input.Enabled,
		})
		return listResult(accounts, total, err)
	}}
}

// productListInput 是商品查询工具输入。
type productListInput struct {
	pagination        // pagination 提供页码与每页数量。
	AccountID  uint   `json:"accountId,omitempty" jsonschema_description:"闲鱼账号 ID"`        // AccountID 限定闲鱼账号。
	TitleOrID  string `json:"titleOrId" jsonschema_description:"商品标题或闲鱼商品 ID 关键词"`         // TitleOrID 按标题或商品标识检索。
	Status     string `json:"status" jsonschema_description:"商品状态：on_sale、offline 或 sold"` // Status 按商品状态筛选。
}

// newProductList 创建只读商品查询工具。
func newProductList(products productReader) *xianyuTool {
	return &xianyuTool{definition: readDefinition(
		"xianyu_product_list", "查询闲鱼商品", "查询当前用户的闲鱼商品及自动回复、发货和续售设置。", productListInput{},
	), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input productListInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		products, total, err := products.List(ctx, scope.OwnerID, xyRequest.ProductSearch{
			Pagination: page(input.pagination), AccountID: input.AccountID,
			TitleOrID: input.TitleOrID, Status: input.Status,
		})
		return listResult(products, total, err)
	}}
}

// orderListInput 是订单查询工具输入。
type orderListInput struct {
	pagination        // pagination 提供页码与每页数量。
	AccountID  uint   `json:"accountId,omitempty" jsonschema_description:"闲鱼账号 ID"` // AccountID 限定闲鱼账号。
	Keyword    string `json:"keyword" jsonschema_description:"订单号、商品标题或买家名称关键词"`    // Keyword 是订单模糊检索词。
	Status     string `json:"status" jsonschema_description:"订单状态"`                 // Status 按订单状态筛选。
}

// newOrderList 创建只读订单查询工具。
func newOrderList(orders orderReader) *xianyuTool {
	return &xianyuTool{definition: readDefinition(
		"xianyu_order_list", "查询闲鱼订单", "查询当前用户的闲鱼订单，不执行同步、发货或状态变更。", orderListInput{},
	), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input orderListInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		orders, total, err := orders.List(ctx, scope.OwnerID, xyRequest.OrderSearch{
			PageInfo:  xyRequest.PageInfo{Pagination: page(input.pagination), Keyword: input.Keyword},
			AccountID: input.AccountID, Status: input.Status,
		})
		return listResult(orders, total, err)
	}}
}

type orderDetailInput struct {
	OrderID uint `json:"orderId" jsonschema_description:"本地订单 ID"` // OrderID 是目标本地订单 ID。
}

func newOrderDetail(orders orderDetailReader) *xianyuTool {
	return &xianyuTool{definition: readDefinition(
		"xianyu_order_detail", "查询闲鱼订单详情", "查询当前用户拥有的单个闲鱼订单详情。", orderDetailInput{},
	), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input orderDetailInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		order, err := orders.Get(ctx, scope.OwnerID, input.OrderID)
		return jsonResult(order, err)
	}}
}

type deliveryTaskListInput struct {
	pagination
	AccountID uint   `json:"accountId,omitempty" jsonschema_description:"闲鱼账号 ID"`    // AccountID 限定账号。
	ProductID uint   `json:"productId,omitempty" jsonschema_description:"本地商品 ID"`    // ProductID 限定商品。
	Status    string `json:"status,omitempty" jsonschema_description:"任务状态"`          // Status 按状态筛选。
	Keyword   string `json:"keyword,omitempty" jsonschema_description:"订单号、商品或买家关键词"` // Keyword 按关联订单信息检索。
}

func newDeliveryTaskList(delivery deliveryReader) *xianyuTool {
	return &xianyuTool{definition: readDefinition(
		"xianyu_delivery_task_list", "查询闲鱼发货任务", "查询当前用户的自动发货任务和失败原因，不会重试任务。", deliveryTaskListInput{},
	), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input deliveryTaskListInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		list, total, err := delivery.ListTasks(ctx, scope.OwnerID, xyRequest.DeliveryTaskSearch{
			PageInfo:  xyRequest.PageInfo{Pagination: page(input.pagination)},
			AccountID: input.AccountID, ProductID: input.ProductID, Status: input.Status, Keyword: input.Keyword,
		})
		return listResult(list, total, err)
	}}
}

type deliveryInventoryInput struct {
	RuleID uint `json:"ruleId" jsonschema_description:"发货规则 ID"` // RuleID 是目标发货规则 ID。
}

func newDeliveryInventory(delivery deliveryReader) *xianyuTool {
	return &xianyuTool{definition: readDefinition(
		"xianyu_delivery_inventory", "查询闲鱼卡密库存", "查询当前用户指定发货规则的可用、预留和已交付卡密数量，不返回卡密内容。", deliveryInventoryInput{},
	), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input deliveryInventoryInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		inventory, err := delivery.Inventory(ctx, scope.OwnerID, input.RuleID)
		return jsonResult(inventory, err)
	}}
}

type replyRuleListInput struct {
	pagination
	ProductID   uint   `json:"productId,omitempty" jsonschema_description:"本地商品 ID"`   // ProductID 限定商品。
	Type        string `json:"type,omitempty" jsonschema_description:"规则类型"`           // Type 按规则类型筛选。
	Keyword     string `json:"keyword,omitempty" jsonschema_description:"关键词或回复内容"`    // Keyword 检索匹配词或回复内容。
	EnabledOnly bool   `json:"enabledOnly,omitempty" jsonschema_description:"只查询启用规则"` // EnabledOnly 仅返回启用规则。
}

func newReplyRuleList(rules replyRuleReader) *xianyuTool {
	return &xianyuTool{definition: readDefinition(
		"xianyu_reply_rule_list", "查询闲鱼自动回复规则", "查询当前用户的自动回复规则及其匹配内容。", replyRuleListInput{},
	), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input replyRuleListInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		list, total, err := rules.List(ctx, scope.OwnerID, xyRequest.ReplyRuleSearch{
			PageInfo:  xyRequest.PageInfo{Pagination: page(input.pagination), Keyword: input.Keyword},
			ProductID: input.ProductID, Type: input.Type, EnabledOnly: input.EnabledOnly,
		})
		return listResult(list, total, err)
	}}
}

type deliveryRuleListInput struct {
	pagination
	RuleID      uint   `json:"ruleId,omitempty" jsonschema_description:"指定发货规则 ID"`     // RuleID 限定规则。
	ProductID   uint   `json:"productId,omitempty" jsonschema_description:"本地商品 ID"`    // ProductID 限定商品。
	Mode        string `json:"mode,omitempty" jsonschema_description:"发货模式"`            // Mode 按发货模式筛选。
	Keyword     string `json:"keyword,omitempty" jsonschema_description:"SKU 或发货模板关键词"` // Keyword 检索 SKU 或模板。
	EnabledOnly bool   `json:"enabledOnly,omitempty" jsonschema_description:"只查询启用规则"`  // EnabledOnly 仅返回启用规则。
}

func newDeliveryRuleList(rules deliveryRuleReader) *xianyuTool {
	return &xianyuTool{definition: readDefinition(
		"xianyu_delivery_rule_list", "查询闲鱼自动发货规则", "查询当前用户的自动发货规则，不返回任何已导入卡密。", deliveryRuleListInput{},
	), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input deliveryRuleListInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		list, total, err := rules.List(ctx, scope.OwnerID, xyRequest.DeliveryRuleSearch{
			PageInfo: xyRequest.PageInfo{Pagination: page(input.pagination), Keyword: input.Keyword},
			RuleID:   input.RuleID, ProductID: input.ProductID, Mode: input.Mode, EnabledOnly: input.EnabledOnly,
		})
		return listResult(list, total, err)
	}}
}

type deliveryImportHandoffInput struct {
	RuleID uint `json:"ruleId,omitempty" jsonschema_description:"可选的闲鱼发货规则 ID"` // RuleID 是可选的目标发货规则。
}

func newDeliveryImportHandoff(rules deliveryRuleReader) *xianyuTool {
	return &xianyuTool{definition: readDefinition("xianyu_delivery_import_handoff", "打开闲鱼卡密导入", "返回受保护的发货规则卡密导入页面；可指定当前用户拥有的卡密规则，不接收或返回卡密内容。", deliveryImportHandoffInput{}), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input deliveryImportHandoffInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		path := "/xianyu/delivery-rules"
		if input.RuleID != 0 {
			_, total, err := rules.List(ctx, scope.OwnerID, xyRequest.DeliveryRuleSearch{
				PageInfo: xyRequest.PageInfo{Pagination: xyRequest.Pagination{Page: 1, PageSize: 1}}, RuleID: input.RuleID,
			})
			if err != nil {
				return "", capabilityError(err)
			}
			if total == 0 {
				return "", &agenttool.ExecutionError{Code: "not_found", Message: "未找到当前用户拥有的发货规则"}
			}
			path += "?importRuleId=" + strconv.FormatUint(uint64(input.RuleID), 10)
		}
		return jsonResult(map[string]string{"path": path}, nil)
	}}
}

type businessEventListInput struct {
	pagination
	AccountID uint   `json:"accountId,omitempty" jsonschema_description:"闲鱼账号 ID"`  // AccountID 限定账号。
	ProductID uint   `json:"productId,omitempty" jsonschema_description:"本地商品 ID"`  // ProductID 限定商品。
	OrderID   uint   `json:"orderId,omitempty" jsonschema_description:"本地订单 ID"`    // OrderID 限定订单。
	Type      string `json:"type,omitempty" jsonschema_description:"事件类型"`          // Type 按业务事件类型筛选。
	Status    string `json:"status,omitempty" jsonschema_description:"事件状态"`        // Status 按执行状态筛选。
	Keyword   string `json:"keyword,omitempty" jsonschema_description:"事件摘要或错误关键词"` // Keyword 检索事件摘要或错误。
}

func newBusinessEventList(events businessEventReader) *xianyuTool {
	return &xianyuTool{definition: readDefinition(
		"xianyu_business_event_list", "查询闲鱼业务事件", "查询当前用户的闲鱼自动化操作记录、结果和安全错误摘要。", businessEventListInput{},
	), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input businessEventListInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		list, total, err := events.List(ctx, scope.OwnerID, xyRequest.BusinessEventSearch{
			PageInfo:  xyRequest.PageInfo{Pagination: page(input.pagination), Keyword: input.Keyword},
			AccountID: input.AccountID, ProductID: input.ProductID, OrderID: input.OrderID,
			Type: input.Type, Status: input.Status,
		})
		return listResult(list, total, err)
	}}
}

type relistTaskListInput struct {
	pagination
	AccountID uint   `json:"accountId,omitempty" jsonschema_description:"闲鱼账号 ID"` // AccountID 限定账号。
	ProductID uint   `json:"productId,omitempty" jsonschema_description:"本地商品 ID"` // ProductID 限定商品。
	Status    string `json:"status,omitempty" jsonschema_description:"任务状态"`       // Status 按状态筛选。
}

func newRelistTaskList(relist relistReader) *xianyuTool {
	return &xianyuTool{definition: readDefinition(
		"xianyu_relist_task_list", "查询闲鱼续售任务", "查询当前用户的自动续售任务状态和安全错误摘要，不执行发布或重试。", relistTaskListInput{},
	), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input relistTaskListInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		list, total, err := relist.ListTasks(ctx, scope.OwnerID, xyRequest.RelistTaskSearch{
			Pagination: page(input.pagination), AccountID: input.AccountID, ProductID: input.ProductID, Status: input.Status,
		})
		return listResult(list, total, err)
	}}
}

// messageListInput 是消息查询工具输入。
type messageListInput struct {
	pagination               // pagination 提供页码与每页数量。
	AccountID         uint   `json:"accountId,omitempty" jsonschema_description:"闲鱼账号 ID"`       // AccountID 限定闲鱼账号。
	SessionID         string `json:"sessionId" jsonschema_description:"闲鱼会话 ID"`                 // SessionID 限定会话。
	Direction         string `json:"direction" jsonschema_description:"消息方向：inbound 或 outbound"` // Direction 限定收发方向。
	ExternalProductID string `json:"externalProductId" jsonschema_description:"闲鱼商品 ID"`         // ExternalProductID 限定关联商品。
	BuyerName         string `json:"buyerName" jsonschema_description:"买家名称关键词"`                 // BuyerName 按买家名称检索。
	Content           string `json:"content" jsonschema_description:"消息内容关键词"`                   // Content 按消息正文检索。
}

// newMessageList 创建只读消息查询工具。
func newMessageList(messages messageService) *xianyuTool {
	return &xianyuTool{definition: readDefinition(
		"xianyu_message_list", "查询闲鱼消息", "查询当前用户的闲鱼会话消息，不发送消息或改变接管状态。", messageListInput{},
	), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input messageListInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		messages, total, err := messages.List(ctx, scope.OwnerID, xyRequest.MessageSearch{
			Pagination: page(input.pagination), AccountID: input.AccountID, SessionID: input.SessionID,
			Direction: input.Direction, ExternalProductID: input.ExternalProductID,
			BuyerName: input.BuyerName, Content: input.Content,
		})
		return listResult(messages, total, err)
	}}
}

// messageReplyInput 是消息回复工具输入。
type messageReplyInput struct {
	MessageID uint   `json:"messageId" jsonschema_description:"要回复的入站消息 ID"` // MessageID 是目标入站消息 ID。
	Text      string `json:"text" jsonschema_description:"回复内容，最多 4000 个字符"` // Text 是待发送的回复正文。
}

// newMessageReply 创建需要用户审批的真实消息发送工具。
func newMessageReply(messages messageService) *xianyuTool {
	definition := agenttool.Definition{
		Code: "xianyu_message_reply", Name: "回复闲鱼消息",
		Description: "回复当前用户拥有的闲鱼消息。该操作会发送真实消息，执行前必须由用户确认。",
		Source:      "local", Parameters: jsonschema.Reflect(messageReplyInput{}),
		RequiresApproval: true, Timeout: 15 * time.Second,
	}
	return &xianyuTool{definition: definition, invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input messageReplyInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		input.Text = strings.TrimSpace(input.Text)
		if input.MessageID == 0 || input.Text == "" || len([]rune(input.Text)) > 4000 {
			return "", &agenttool.ExecutionError{Code: "invalid_arguments", Message: "消息 ID 和回复内容不能为空，内容不能超过 4000 个字符"}
		}
		if err := messages.Reply(ctx, scope.OwnerID, xyRequest.SendMessage{MessageID: input.MessageID, Text: input.Text}); err != nil {
			return "", capabilityError(err)
		}
		return `{"ok":true}`, nil
	}}
}

type accountSyncInput struct {
	AccountID uint `json:"accountId" jsonschema_description:"闲鱼账号 ID"` // AccountID 是同步目标账号。
}

func newProductSync(products productWriter) *xianyuTool {
	return &xianyuTool{definition: writeDefinition("xianyu_product_sync", "同步闲鱼商品", "同步当前用户指定账号的闲鱼商品及自动化设置。", accountSyncInput{}), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input accountSyncInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		count, err := products.Sync(ctx, scope.OwnerID, xyRequest.ProductSync{AccountID: input.AccountID})
		return jsonResult(map[string]any{"ok": err == nil, "count": count}, err)
	}}
}

func newOrderSync(orders orderWriter) *xianyuTool {
	return &xianyuTool{definition: writeDefinition("xianyu_pending_order_sync", "同步待发货订单", "同步当前用户指定账号的待发货订单，并按现有规则规划自动发货。", accountSyncInput{}), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input accountSyncInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		count, err := orders.SyncPending(ctx, scope.OwnerID, input.AccountID)
		return jsonResult(map[string]any{"ok": err == nil, "count": count}, err)
	}}
}

func newProductSettingPatch(products productWriter) *xianyuTool {
	return &xianyuTool{definition: writeDefinition("xianyu_product_setting_patch", "调整闲鱼商品自动化", "只修改明确提供的自动回复、自动发货、自动续售及其参数，未提及的设置保持不变。", xyRequest.PatchProductSetting{}), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input xyRequest.PatchProductSetting
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		if err := products.PatchSetting(ctx, scope.OwnerID, input); err != nil {
			return "", capabilityError(err)
		}
		return `{"ok":true}`, nil
	}}
}

type replyRuleSaveInput struct {
	ID        uint   `json:"id,omitempty" jsonschema_description:"更新时填写规则 ID，新建时省略"` // ID 是更新规则的本地标识。
	ProductID uint   `json:"productId"`                                              // ProductID 是规则绑定商品。
	Type      string `json:"type"`                                                   // Type 是精确、模糊或兜底匹配类型。
	Keyword   string `json:"keyword,omitempty"`                                      // Keyword 是匹配关键词。
	Priority  int    `json:"priority,omitempty"`                                     // Priority 是规则匹配顺序。
	Enabled   bool   `json:"enabled"`                                                // Enabled 决定规则是否启用。
	Content   string `json:"content"`                                                // Content 是自动回复文本。
}

func newReplyRuleSave(rules replyRuleWriter) *xianyuTool {
	return &xianyuTool{definition: writeDefinition("xianyu_reply_rule_save", "保存闲鱼自动回复规则", "新建或更新当前用户商品的自动回复规则。", replyRuleSaveInput{}), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input replyRuleSaveInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		err := rules.Save(ctx, scope.OwnerID, xyRequest.SaveReplyRule{ID: input.ID, ProductID: input.ProductID, Type: input.Type, Keyword: input.Keyword, Priority: input.Priority, Enabled: input.Enabled, Content: input.Content})
		return jsonResult(map[string]bool{"ok": err == nil}, err)
	}}
}

type ruleIDInput struct {
	ID uint `json:"id" jsonschema_description:"目标规则 ID"` // ID 是目标规则本地标识。
}

func newReplyRuleDelete(rules replyRuleWriter) *xianyuTool {
	return &xianyuTool{definition: writeDefinition("xianyu_reply_rule_delete", "删除闲鱼自动回复规则", "删除当前用户拥有的自动回复规则。", ruleIDInput{}), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input ruleIDInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		err := rules.Delete(ctx, scope.OwnerID, input.ID)
		return jsonResult(map[string]bool{"ok": err == nil}, err)
	}}
}

type deliveryRuleSaveInput struct {
	ID        uint   `json:"id,omitempty" jsonschema_description:"更新时填写规则 ID，新建时省略"` // ID 是更新规则的本地标识。
	ProductID uint   `json:"productId"`                                              // ProductID 是规则绑定商品。
	SKUKey    string `json:"skuKey,omitempty"`                                       // SKUKey 是商品规格标识。
	SKUName   string `json:"skuName,omitempty"`                                      // SKUName 是商品规格名称。
	Mode      string `json:"mode"`                                                   // Mode 是文本或卡密发货模式。
	Template  string `json:"template"`                                               // Template 是发货消息模板，不是卡密内容。
	Enabled   bool   `json:"enabled"`                                                // Enabled 决定规则是否启用。
}

func newDeliveryRuleSave(rules deliveryRuleWriter) *xianyuTool {
	return &xianyuTool{definition: writeDefinition("xianyu_delivery_rule_save", "保存闲鱼自动发货规则", "新建或更新当前用户商品的自动发货规则；卡密内容必须通过受保护导入界面提交。", deliveryRuleSaveInput{}), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input deliveryRuleSaveInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		err := rules.Save(ctx, scope.OwnerID, xyRequest.SaveDeliveryRule{ID: input.ID, ProductID: input.ProductID, SKUKey: input.SKUKey, SKUName: input.SKUName, Mode: input.Mode, Template: input.Template, Enabled: input.Enabled})
		return jsonResult(map[string]bool{"ok": err == nil}, err)
	}}
}

func newDeliveryRuleDelete(rules deliveryRuleWriter) *xianyuTool {
	return &xianyuTool{definition: writeDefinition("xianyu_delivery_rule_delete", "删除闲鱼自动发货规则", "删除当前用户拥有的自动发货规则。", ruleIDInput{}), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input ruleIDInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		err := rules.Delete(ctx, scope.OwnerID, input.ID)
		return jsonResult(map[string]bool{"ok": err == nil}, err)
	}}
}

func newDeliveryRetry(delivery deliveryWriter) *xianyuTool {
	return &xianyuTool{definition: writeDefinition("xianyu_delivery_retry", "重试闲鱼发货任务", "重试当前用户明确失败的自动发货任务。", retryTaskInput{}), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input retryTaskInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		err := delivery.Retry(ctx, scope.OwnerID, input.TaskID)
		return jsonResult(map[string]bool{"ok": err == nil}, err)
	}}
}

func newRelistRetry(relist relistWriter) *xianyuTool {
	return &xianyuTool{definition: writeDefinition("xianyu_relist_retry", "重试闲鱼续售任务", "仅重试当前用户状态明确为 failed 的续售任务；unknown 状态不可重试。", retryTaskInput{}), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input retryTaskInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		err := relist.Retry(ctx, scope.OwnerID, input.TaskID)
		return jsonResult(map[string]bool{"ok": err == nil}, err)
	}}
}

type retryTaskInput struct {
	TaskID uint `json:"taskId" jsonschema_description:"目标任务 ID"` // TaskID 是待重试任务标识。
}

type takeoverInput struct {
	MessageID uint `json:"messageId" jsonschema_description:"定位会话的消息 ID"` // MessageID 是目标会话消息。
	Active    bool `json:"active"`                                        // Active 决定是否启用人工接管。
}

func newMessageTakeover(messages messageService) *xianyuTool {
	return &xianyuTool{definition: writeDefinition("xianyu_message_takeover", "设置闲鱼会话接管", "开启或关闭当前用户闲鱼会话的人工接管状态。", takeoverInput{}), invoke: func(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
		var input takeoverInput
		if err := decode(arguments, &input); err != nil {
			return "", err
		}
		err := messages.UpdateTakeover(ctx, scope.OwnerID, xyRequest.UpdateTakeover{MessageID: input.MessageID, Active: input.Active})
		return jsonResult(map[string]bool{"ok": err == nil}, err)
	}}
}

// readDefinition 创建本地只读工具的统一定义。
func readDefinition(code, name, description string, input any) agenttool.Definition {
	return agenttool.Definition{
		Code: code, Name: name, Description: description, Source: "local",
		Parameters: jsonschema.Reflect(input), ReadOnly: true, Timeout: 10 * time.Second,
	}
}

func writeDefinition(code, name, description string, input any) agenttool.Definition {
	definition := readDefinition(code, name, description, input)
	definition.ReadOnly = false
	definition.RequiresApproval = true
	return definition
}

// page 将工具分页输入转换为服务契约。
func page(input pagination) xyRequest.Pagination {
	return xyRequest.Pagination{Page: input.Page, PageSize: input.PageSize}
}

// decode 解码工具参数并转换为稳定执行错误。
func decode(arguments json.RawMessage, input any) error {
	if err := json.Unmarshal(arguments, input); err != nil {
		return &agenttool.ExecutionError{Code: "invalid_arguments", Message: "工具参数无效", Cause: err}
	}
	return nil
}

// listResult 将分页服务结果编码为工具 JSON 文本。
func listResult(list any, total int64, err error) (string, error) {
	if err != nil {
		return "", capabilityError(err)
	}
	payload, err := json.Marshal(map[string]any{"list": list, "total": total})
	if err != nil {
		return "", &agenttool.ExecutionError{Code: "invalid_result", Message: "闲鱼能力返回结果无效", Cause: err}
	}
	return string(payload), nil
}

func jsonResult(value any, err error) (string, error) {
	if err != nil {
		return "", capabilityError(err)
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return "", &agenttool.ExecutionError{Code: "invalid_result", Message: "闲鱼能力返回结果无效", Cause: err}
	}
	return string(payload), nil
}

// capabilityError 隐藏服务实现细节并保留原错误链。
func capabilityError(err error) error {
	return &agenttool.ExecutionError{Code: "xianyu_operation_failed", Message: "闲鱼业务能力执行失败", Cause: fmt.Errorf("xianyu service: %w", err)}
}
