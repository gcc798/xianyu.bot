package xianyutool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/eino-contrib/jsonschema"
	agenttool "github.com/gcc798/xianyu.bot/internal/agent/tool"
	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/response"
)

type accountReader interface {
	List(context.Context, uint, xyRequest.AccountSearch) ([]model.Account, int64, error)
}

type productReader interface {
	List(context.Context, uint, xyRequest.ProductSearch) ([]xyResponse.Product, int64, error)
}

type orderReader interface {
	List(context.Context, uint, xyRequest.OrderSearch) ([]xyResponse.Order, int64, error)
}

type messageService interface {
	List(context.Context, uint, xyRequest.MessageSearch) ([]xyResponse.Message, int64, error)
	Reply(context.Context, uint, xyRequest.SendMessage) error
}

type xianyuTool struct {
	definition agenttool.Definition
	invoke     func(context.Context, agenttool.Scope, json.RawMessage) (string, error)
}

func (t *xianyuTool) Definition() agenttool.Definition { return t.definition }

func (t *xianyuTool) Invoke(ctx context.Context, scope agenttool.Scope, arguments json.RawMessage) (string, error) {
	return t.invoke(ctx, scope, arguments)
}

func Register(registry *agenttool.Registry, accounts accountReader, products productReader, orders orderReader, messages messageService) error {
	if registry == nil || accounts == nil || products == nil || orders == nil || messages == nil {
		return errors.New("闲鱼工具依赖不能为空")
	}
	tools := []*xianyuTool{
		newAccountList(accounts), newProductList(products), newOrderList(orders),
		newMessageList(messages), newMessageReply(messages),
	}
	for _, implementation := range tools {
		if err := registry.Register(implementation); err != nil {
			return err
		}
	}
	return nil
}

type pagination struct {
	Page     int `json:"page" jsonschema_description:"页码，默认 1"`
	PageSize int `json:"pageSize" jsonschema_description:"每页数量，最大 100，默认 10"`
}

type accountListInput struct {
	pagination
	Keyword string `json:"keyword" jsonschema_description:"账号标签、备注或闲鱼用户 ID 关键词"`
	Status  string `json:"status" jsonschema_description:"账号状态"`
	Enabled *bool  `json:"enabled,omitempty" jsonschema_description:"是否启用"`
}

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

type productListInput struct {
	pagination
	AccountID uint   `json:"accountId,omitempty" jsonschema_description:"闲鱼账号 ID"`
	TitleOrID string `json:"titleOrId" jsonschema_description:"商品标题或闲鱼商品 ID 关键词"`
	Status    string `json:"status" jsonschema_description:"商品状态：on_sale、offline 或 sold"`
}

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

type orderListInput struct {
	pagination
	AccountID uint   `json:"accountId,omitempty" jsonschema_description:"闲鱼账号 ID"`
	Keyword   string `json:"keyword" jsonschema_description:"订单号、商品标题或买家名称关键词"`
	Status    string `json:"status" jsonschema_description:"订单状态"`
}

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

type messageListInput struct {
	pagination
	AccountID         uint   `json:"accountId,omitempty" jsonschema_description:"闲鱼账号 ID"`
	SessionID         string `json:"sessionId" jsonschema_description:"闲鱼会话 ID"`
	Direction         string `json:"direction" jsonschema_description:"消息方向：inbound 或 outbound"`
	ExternalProductID string `json:"externalProductId" jsonschema_description:"闲鱼商品 ID"`
	BuyerName         string `json:"buyerName" jsonschema_description:"买家名称关键词"`
	Content           string `json:"content" jsonschema_description:"消息内容关键词"`
}

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

type messageReplyInput struct {
	MessageID uint   `json:"messageId" jsonschema_description:"要回复的入站消息 ID"`
	Text      string `json:"text" jsonschema_description:"回复内容，最多 4000 个字符"`
}

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

func readDefinition(code, name, description string, input any) agenttool.Definition {
	return agenttool.Definition{
		Code: code, Name: name, Description: description, Source: "local",
		Parameters: jsonschema.Reflect(input), ReadOnly: true, Timeout: 10 * time.Second,
	}
}

func page(input pagination) xyRequest.Pagination {
	return xyRequest.Pagination{Page: input.Page, PageSize: input.PageSize}
}

func decode(arguments json.RawMessage, input any) error {
	if err := json.Unmarshal(arguments, input); err != nil {
		return &agenttool.ExecutionError{Code: "invalid_arguments", Message: "工具参数无效", Cause: err}
	}
	return nil
}

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

func capabilityError(err error) error {
	return &agenttool.ExecutionError{Code: "xianyu_operation_failed", Message: "闲鱼业务能力执行失败", Cause: fmt.Errorf("xianyu service: %w", err)}
}
