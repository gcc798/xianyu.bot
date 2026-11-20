# Proposal

## Why

当前 Agent 每次运行都会串行连接所有启用的 MCP、重新获取完整工具目录并把全部工具定义交给模型；当 MCP 和 Tool 数量增长时，这会放大启动延迟、模型上下文、工具误选和故障等待。系统需要把 MCP 能力目录、运行时连接和单轮工具选择拆开，既支持管理员判断服务健康，也让 Agent 只加载当前任务相关的工具。

## What Changes

- 持久化 MCP Tool Catalog，并在进程内维护可原子替换的热目录；后台以有界并发和 TTL 刷新目录，支持分页、目录哈希和失败时保留最近成功快照。
- 增加独立 MCP 健康检查，记录最近状态、检查时间、响应耗时和安全错误摘要；健康检查不刷新工具目录、不延长执行连接 TTL，也不产生 Agent Tool 审计。
- 将 MCP 工具定义与活动连接解耦；实际调用时建立并复用每个 MCP 的共享客户端，允许并发 Tool Call，不限制活跃 MCP 总数，只关闭超过空闲 TTL 且没有在途调用的连接。
- 增加 MCP 能力描述、标签和 Tool allow/deny 配置；使用权限过滤、MCP 选择和 Tool 选择两级路由，默认缩小单步活跃工具集，低置信度和复杂任务可以扩大候选。
- 保留现有 Tool 预算、审批、Owner Scope、参数过滤和审计；审批执行前校验工具定义快照，定义变化时要求重新提议。
- 扩展 MCP 管理 API 和设置界面，分别展示健康状态、目录状态、工具数量和最近错误，并提供手动健康检查与目录刷新。
- 更新架构和产品文档；本次不引入向量数据库，只有真实路由评估证明需要时才考虑 PostgreSQL `pgvector`。

## Capabilities

### New Capabilities

- `mcp-capability-catalog`: 持久化、刷新和查询 MCP 服务及 Tool Catalog，并以两级路由向 Agent 提供当前任务相关工具。
- `mcp-connection-health`: 独立检查 MCP 协议可用性，并安全复用、并发使用和按空闲 TTL 回收实际 Tool 调用连接。

### Modified Capabilities

None. The repository currently has no published OpenSpec capabilities.

## Impact

- 后端 Agent/MCP：`server/internal/agent` 的注册表、Runner、MCP Tool、审批、资源任务和运行审计。
- 数据模型和初始化结构：`server/internal/model` 与 `server/internal/infra/migrations/init.sql`，不新增物理外键、CHECK、ENUM 或触发器。
- MCP 管理 API：`server/internal/agent/ai_config.go`、`server/internal/api` 与路由。
- 控制台：`agent-web` 的 MCP 类型、API 和 `MCPSettings`。
- 启动组装：`server/cmd/server/main.go` 中后台刷新、健康检查和连接回收生命周期。
- 文档与验证：Agent/MCP 架构文档、Go 测试、前端构建、浏览器检查和受影响容器验证。
