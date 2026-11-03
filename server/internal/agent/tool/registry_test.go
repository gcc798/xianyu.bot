package agenttool

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/eino-contrib/jsonschema"
	"github.com/stretchr/testify/require"
)

type stubTool struct {
	definition Definition
}

func definition(code, name string, parameters *jsonschema.Schema) Definition {
	return Definition{
		Code: code, Name: name, Description: "测试工具", Source: "local",
		Parameters: parameters, Timeout: time.Second,
	}
}

func (s stubTool) Definition() Definition { return s.definition }

func (stubTool) Invoke(context.Context, Scope, json.RawMessage) (string, error) {
	return "ok", nil
}

func TestRegistryRejectsInvalidAndDuplicateCodes(t *testing.T) {
	registry := NewRegistry()
	parameters := &jsonschema.Schema{Type: "object"}
	require.Error(t, registry.Register(stubTool{definition: definition("", "无编码", parameters)}))
	require.NoError(t, registry.Register(stubTool{definition: definition("xianyu_order_list", "查询订单", parameters)}))
	require.Error(t, registry.Register(stubTool{definition: definition("xianyu_order_list", "重复订单", parameters)}))
}

func TestRegistryListsToolsByCode(t *testing.T) {
	registry := NewRegistry()
	parameters := &jsonschema.Schema{Type: "object"}
	require.NoError(t, registry.Register(stubTool{definition: definition("xianyu_product_list", "查询商品", parameters)}))
	require.NoError(t, registry.Register(stubTool{definition: definition("xianyu_account_list", "查询账号", parameters)}))

	tools := registry.List()
	require.Len(t, tools, 2)
	require.Equal(t, "xianyu_account_list", tools[0].Definition().Code)
	require.Equal(t, "xianyu_product_list", tools[1].Definition().Code)
	registered, exists := registry.Get("xianyu_account_list")
	require.True(t, exists)
	require.Equal(t, "查询账号", registered.Definition().Name)
}
