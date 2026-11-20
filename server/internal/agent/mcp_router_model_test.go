package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/stretchr/testify/require"
)

// TestRouteMCPToolsUsesStructuredTwoLevelSelection 验证两级编码选择和模型预算。
func TestRouteMCPToolsUsesStructuredTwoLevelSelection(t *testing.T) {
	registry := routingFixture(3, 10)
	chat := &routeTestModel{responses: []string{
		`{"mcp_codes":["mcp_1"],"confidence":0.9}`,
		`{"tool_codes":["mcp_1__tool_3"],"confidence":0.9}`,
	}}
	harness := newAgentHarness(model.DefaultAgentConfig())
	route, err := routeMCPTools(t.Context(), chat, harness, registry, 7, "处理这个未明确领域的任务", nil)
	require.NoError(t, err)
	require.Equal(t, "model_two_level", route.Strategy)
	require.Equal(t, []string{"mcp_1__tool_3"}, route.Codes)
	require.Equal(t, []string{"mcp_1"}, route.MCPs)
	require.Len(t, route.SearchTools, 9)
	require.NotContains(t, routeToolCodes(route.SearchTools), "mcp_1__tool_3")
	require.Equal(t, 2, harness.Usage().ModelCalls)
}

// TestDynamicMCPToolsRemainGoverned 验证运行中可搜索工具仍逐个经过统一治理适配器。
func TestDynamicMCPToolsRemainGoverned(t *testing.T) {
	registry := routingFixture(3, 10)
	route, err := routeMCPTools(t.Context(), new(routeTestModel), newAgentHarness(model.DefaultAgentConfig()), registry, 7, "使用 mcp_1__tool_3", nil)
	require.NoError(t, err)
	require.Len(t, route.SearchTools, 9)

	tools, err := buildEinoMCPTools(nil, registry, route.SearchTools, newAgentHarness(model.DefaultAgentConfig()), nil, model.AgentRun{OwnerID: 7})
	require.NoError(t, err)
	require.Len(t, tools, 9)
	for _, candidate := range tools {
		governed, ok := candidate.(*governedTool)
		require.True(t, ok)
		_, ok = governed.implementation.(*mcpAgentTool)
		require.True(t, ok)
	}
}

// TestRouteMCPToolsExpandsOnUnknownAndComplexSelection 验证未知编码拒绝和低置信度/复杂任务扩展。
func TestRouteMCPToolsExpandsOnUnknownAndComplexSelection(t *testing.T) {
	registry := routingFixture(3, 10)
	unknown := &routeTestModel{responses: []string{`{"mcp_codes":["invented"],"confidence":0.9}`}}
	route, err := routeMCPTools(t.Context(), unknown, newAgentHarness(model.DefaultAgentConfig()), registry, 7, "处理这个任务", nil)
	require.NoError(t, err)
	require.True(t, route.Expanded)
	require.Equal(t, "expanded_fallback", route.Strategy)
	require.Len(t, route.Tools, 30)
	require.NotContains(t, route.Codes, "invented")

	complex := &routeTestModel{responses: []string{
		`{"mcp_codes":["mcp_1"],"confidence":0.9}`,
		`{"tool_codes":["mcp_1__tool_3"],"confidence":0.9}`,
	}}
	route, err = routeMCPTools(t.Context(), complex, newAgentHarness(model.DefaultAgentConfig()), registry, 7, "先处理资料，同时发送结果，然后继续检查", nil)
	require.NoError(t, err)
	require.True(t, route.Expanded)
	require.Greater(t, len(route.Tools), 1)
	require.LessOrEqual(t, len(route.Tools), expandedMCPToolLimit)
}

// TestRouteMCPToolsSkipsModelForExplicitMatchAndEnforcesBudget 验证明确定性路由和路由模型预算。
func TestRouteMCPToolsSkipsModelForExplicitMatchAndEnforcesBudget(t *testing.T) {
	registry := routingFixture(3, 10)
	harness := newAgentHarness(model.DefaultAgentConfig())
	route, err := routeMCPTools(t.Context(), new(routeTestModel), harness, registry, 7, "使用 mcp_1__tool_3", nil)
	require.NoError(t, err)
	require.Equal(t, "deterministic", route.Strategy)
	require.Equal(t, []string{"mcp_1__tool_3"}, route.Codes)
	require.Zero(t, harness.Usage().ModelCalls)

	limitedConfig := model.DefaultAgentConfig()
	limitedConfig.MaxModelCalls = 1
	limited := newAgentHarness(limitedConfig)
	chat := &routeTestModel{responses: []string{`{"mcp_codes":["mcp_1"],"confidence":0.9}`}}
	_, err = routeMCPTools(t.Context(), chat, limited, registry, 7, "处理这个任务", nil)
	require.ErrorIs(t, err, ErrAgentBudgetExceeded)
	require.Equal(t, 1, limited.Usage().ModelCalls)
}

// routingFixture 创建多个 MCP 和工具的内存路由目录。
func routingFixture(serverCount, toolsPerServer int) *MCPRegistry {
	servers := make([]model.MCPServer, 0, serverCount)
	tools := make([]model.MCPTool, 0, serverCount*toolsPerServer)
	now := time.Now()
	for serverIndex := range serverCount {
		serverCode := fmt.Sprintf("mcp_%d", serverIndex)
		servers = append(servers, model.MCPServer{Name: "领域", Code: serverCode, Description: fmt.Sprintf("domain-%d", serverIndex), Enabled: true})
		for toolIndex := range toolsPerServer {
			name := fmt.Sprintf("tool_%d", toolIndex)
			definition := fmt.Sprintf(`{"name":%q,"description":%q,"inputSchema":{"type":"object"}}`, name, fmt.Sprintf("capability-%d-%d", serverIndex, toolIndex))
			tools = append(tools, catalogModelTool(serverCode, name, definition, now))
		}
	}
	registry := NewMCPRegistry(nil)
	registry.catalog.Store(buildMCPCatalog(servers, tools))
	return registry
}

// routeTestModel 按顺序返回结构化路由结果。
type routeTestModel struct {
	mu        sync.Mutex
	responses []string
}

func (m *routeTestModel) Generate(context.Context, []*schema.Message, ...einomodel.Option) (*schema.Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.responses) == 0 {
		return nil, errors.New("unexpected route model call")
	}
	response := m.responses[0]
	m.responses = m.responses[1:]
	return schema.AssistantMessage(response, nil), nil
}

func (*routeTestModel) Stream(context.Context, []*schema.Message, ...einomodel.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, errors.New("unexpected route stream call")
}
