package xianyutool

import (
	"context"
	"testing"

	agenttool "github.com/gcc798/xianyu.bot/internal/agent/tool"
	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/stretchr/testify/require"
)

type accountReaderStub struct{ ownerID uint }

func (s *accountReaderStub) List(_ context.Context, ownerID uint, _ xyRequest.AccountSearch) ([]model.Account, int64, error) {
	s.ownerID = ownerID
	return []model.Account{{Label: "主账号"}}, 1, nil
}

type productReaderStub struct{}

func (*productReaderStub) List(context.Context, uint, xyRequest.ProductSearch) ([]xyResponse.Product, int64, error) {
	return nil, 0, nil
}

type orderReaderStub struct{}

func (*orderReaderStub) List(context.Context, uint, xyRequest.OrderSearch) ([]xyResponse.Order, int64, error) {
	return nil, 0, nil
}

type messageServiceStub struct {
	ownerID uint
	reply   xyRequest.SendMessage
}

func (*messageServiceStub) List(context.Context, uint, xyRequest.MessageSearch) ([]xyResponse.Message, int64, error) {
	return nil, 0, nil
}

func (s *messageServiceStub) Reply(_ context.Context, ownerID uint, input xyRequest.SendMessage) error {
	s.ownerID, s.reply = ownerID, input
	return nil
}

func TestRegisterAddsStableXianyuToolDefinitions(t *testing.T) {
	registry := agenttool.NewRegistry()
	require.NoError(t, Register(registry, new(accountReaderStub), new(productReaderStub), new(orderReaderStub), new(messageServiceStub)))

	tools := registry.List()
	require.Len(t, tools, 5)
	require.Equal(t, "xianyu_account_list", tools[0].Definition().Code)
	reply, exists := registry.Get("xianyu_message_reply")
	require.True(t, exists)
	require.Equal(t, "回复闲鱼消息", reply.Definition().Name)
	require.True(t, reply.Definition().RequiresApproval)
	require.False(t, reply.Definition().ReadOnly)
}

func TestAccountToolUsesAuthenticatedOwner(t *testing.T) {
	accounts := new(accountReaderStub)
	registry := agenttool.NewRegistry()
	require.NoError(t, Register(registry, accounts, new(productReaderStub), new(orderReaderStub), new(messageServiceStub)))

	tool, exists := registry.Get("xianyu_account_list")
	require.True(t, exists)
	output, err := tool.Invoke(t.Context(), agenttool.Scope{OwnerID: 42}, []byte(`{"page":1,"pageSize":5}`))
	require.NoError(t, err)
	require.EqualValues(t, 42, accounts.ownerID)
	require.JSONEq(t, `{"list":[{"ID":0,"CreatedAt":"0001-01-01T00:00:00Z","UpdatedAt":"0001-01-01T00:00:00Z","externalUserId":"","label":"主账号","note":"","status":"","enabled":false,"hasCredential":false}],"total":1}`, output)
}

func TestMessageReplyToolDoesNotAcceptOwnerFromArguments(t *testing.T) {
	messages := new(messageServiceStub)
	tool := newMessageReply(messages)
	_, err := tool.Invoke(t.Context(), agenttool.Scope{OwnerID: 7}, []byte(`{"ownerId":99,"messageId":12,"text":" 已处理 "}`))
	require.NoError(t, err)
	require.EqualValues(t, 7, messages.ownerID)
	require.EqualValues(t, 12, messages.reply.MessageID)
	require.Equal(t, "已处理", messages.reply.Text)
}
