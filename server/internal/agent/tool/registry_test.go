package agenttool

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/eino-contrib/jsonschema"
	"github.com/stretchr/testify/require"
)

type stubTool struct {
	definition Definition
}

func (s stubTool) Definition() Definition { return s.definition }

func (stubTool) Invoke(context.Context, Scope, json.RawMessage) (string, error) {
	return "ok", nil
}

func TestRegistryRejectsInvalidAndDuplicateCodes(t *testing.T) {
	registry := NewRegistry()
	parameters := &jsonschema.Schema{Type: "object"}
	require.Error(t, registry.Register(stubTool{definition: Definition{Code: "", Name: "无编码", Parameters: parameters}}))
	require.NoError(t, registry.Register(stubTool{definition: Definition{Code: "xianyu_order_list", Name: "查询订单", Parameters: parameters}}))
	require.Error(t, registry.Register(stubTool{definition: Definition{Code: "xianyu_order_list", Name: "重复订单", Parameters: parameters}}))
}

func TestRegistryListsToolsByCode(t *testing.T) {
	registry := NewRegistry()
	parameters := &jsonschema.Schema{Type: "object"}
	require.NoError(t, registry.Register(stubTool{definition: Definition{Code: "xianyu_product_list", Name: "查询商品", Parameters: parameters}}))
	require.NoError(t, registry.Register(stubTool{definition: Definition{Code: "xianyu_account_list", Name: "查询账号", Parameters: parameters}}))

	tools := registry.List()
	require.Len(t, tools, 2)
	require.Equal(t, "xianyu_account_list", tools[0].Definition().Code)
	require.Equal(t, "xianyu_product_list", tools[1].Definition().Code)
	registered, exists := registry.Get("xianyu_account_list")
	require.True(t, exists)
	require.Equal(t, "查询账号", registered.Definition().Name)
}
