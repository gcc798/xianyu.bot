package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/stretchr/testify/require"
)

// TestMCPRoutingScaleRecallAndSchemaReduction 验证千工具目录的并发召回、零连接路由和 Schema 缩减。
func TestMCPRoutingScaleRecallAndSchemaReduction(t *testing.T) {
	registry := scaleRoutingFixture(100, 10)
	var connections atomic.Int64
	registry.runtimeConnect = func(context.Context, model.MCPServer) (connectedMCPClient, error) {
		connections.Add(1)
		return nil, fmt.Errorf("路由阶段不应连接 MCP")
	}

	type result struct {
		mcpHit  bool
		toolHit bool
		route   mcpToolRoute
		err     error
	}
	results := make(chan result, 100)
	var group sync.WaitGroup
	for index := range 100 {
		group.Add(1)
		go func() {
			defer group.Done()
			code := fmt.Sprintf("domain_%d__action_7", index)
			route, err := routeMCPTools(t.Context(), new(routeTestModel), newAgentHarness(model.DefaultAgentConfig()), registry, 7, "请使用 "+code+" 完成任务", nil)
			results <- result{
				mcpHit:  slices.Contains(route.MCPs, fmt.Sprintf("domain_%d", index)),
				toolHit: slices.Contains(route.Codes, code), route: route, err: err,
			}
		}()
	}
	group.Wait()
	close(results)

	mcpHits, toolHits := 0, 0
	var sample mcpToolRoute
	for current := range results {
		require.NoError(t, current.err)
		require.LessOrEqual(t, len(current.route.MCPs), defaultMCPServerTarget)
		require.LessOrEqual(t, len(current.route.Tools), defaultMCPToolTarget)
		if current.mcpHit {
			mcpHits++
		}
		if current.toolHit {
			toolHits++
		}
		sample = current.route
	}
	require.GreaterOrEqual(t, float64(mcpHits)/100, 0.95, "MCP Recall@5")
	require.GreaterOrEqual(t, float64(toolHits)/100, 0.95, "Tool Recall@20")
	require.Zero(t, connections.Load())

	fullSchemas, err := json.Marshal(registry.CatalogTools())
	require.NoError(t, err)
	routedSchemas, err := json.Marshal(sample.Tools)
	require.NoError(t, err)
	require.Less(t, len(routedSchemas)*10, len(fullSchemas), "单步 Schema 应小于全量目录的 10%")
}

func scaleRoutingFixture(serverCount, toolsPerServer int) *MCPRegistry {
	servers := make([]model.MCPServer, 0, serverCount)
	tools := make([]model.MCPTool, 0, serverCount*toolsPerServer)
	now := time.Now()
	for serverIndex := range serverCount {
		serverCode := fmt.Sprintf("domain_%d", serverIndex)
		servers = append(servers, model.MCPServer{Name: fmt.Sprintf("领域 %d", serverIndex), Code: serverCode, Description: fmt.Sprintf("domain capability %d", serverIndex), Enabled: true})
		for toolIndex := range toolsPerServer {
			name := fmt.Sprintf("action_%d", toolIndex)
			definition := fmt.Sprintf(`{"name":%q,"description":%q,"inputSchema":{"type":"object","properties":{"query":{"type":"string","description":"search terms"},"limit":{"type":"integer"},"cursor":{"type":"string"}},"required":["query"]}}`, name, fmt.Sprintf("domain %d action %d", serverIndex, toolIndex))
			tools = append(tools, catalogModelTool(serverCode, name, definition, now))
		}
	}
	registry := NewMCPRegistry(nil)
	registry.catalog.Store(buildMCPCatalog(servers, tools))
	return registry
}
