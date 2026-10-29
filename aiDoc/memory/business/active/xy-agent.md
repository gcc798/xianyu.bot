# XyAgent

## 基本信息

- 提出日期：2026-10-29
- 当前状态：`active`
- 需求类型：AI Agent 与 MCP 能力平台
- 优先级：P0
- 本地需求文档：`requirements/xy-agent-development-requirements.md`（该目录按用户要求不纳入 Git）

## 用户原始意图摘要

在 GVA 中实现通用的 XyAgent。首个业务能力是通过 Telegram MCP 完成资源搜索、候选检查、异步获取、RustFS 存储和本地元数据管理；后续可以继续接入其他 MCP 能力。Agent 需要 Eino Loop、Skill、Harness、SSE、多轮会话和 `/skill`、`/tool` 手动能力选择。

## 已确认约束

- 所有新增数据库表、代码、路由和前端标识统一使用 `xy` / `Xy`；历史 `xianyu/Xianyu` 标识通过独立迁移阶段处理。
- Agent 名称为 `XyAgent`，不能以资源检索命名。
- MCP 固定提供渠道信息、渠道状态、来源列表、来源同步、资源搜索、资源检查、资源获取、`task_get` 和 `task_cancel`。
- MCP 能力不区分版本；未实现时返回 `not_supported`。
- 不实现 MCP 工具白名单；启用的 MCP Server 返回的全部工具均可供 Agent 使用。
- AI 模型和 MCP 配置保存在数据库，API Key 和 Token 按用户要求明文存储，但不得进入响应、日志、SSE 或模型上下文。
- 新增一级 AI 菜单，子菜单为 Agent、基础配置；基础配置包含模型、MCP 服务和 Agent 配置。
- `/` 默认展示全部能力，`/skill` 和 `/tool` 分类筛选，能力类型使用颜色、图标和文字共同区分。
- 普通对话的不可用提示不暴露 MCP 或具体渠道连接细节。
- 不实现配置修改日志。
- 用户会话、任务和资源使用 `owner_id` 隔离；资源元数据不写商品表。
- 渠道负责获取文件，GVA 负责一次性上传授权、RustFS 控制和最终业务落库。

## 影响范围

- 后端：`server/plugin/xy/`、Agent/MCP/资源模型、API、Service、Router、初始化和权限。
- 前端：`web/src/plugin/xy/`、AI 菜单、Agent 对话页、基础配置页和 SSE 客户端。
- Telegram：现有资源检索拆分、MCP Server、通用任务和交付流程。
- 部署：模型/MCP 数据库配置、服务间认证、RustFS Upload Broker。
- 迁移：历史 `xianyu/Xianyu` 目录、标识、路由、菜单、Casbin、环境变量和测试。

## 当前状态

- 完成需求和架构讨论。
- 已生成本地开发需求文档。
- 已完成 `xy/Xy` 命名迁移，并保留外部协议字面量和既有数据库实例名的兼容性。
- 已完成数据库驱动的模型、MCP 服务和 Agent 运行限制配置，以及 AI 菜单和基础配置页面。
- Telegram 已通过 Streamable HTTP 暴露固定的 9 个 MCP 工具；来源与任务能力已接通，资源搜索、检查和获取暂返回 `not_supported`。
- 待完成：资源搜索/获取拆分、GVA 交付与资源表、Eino Agent Loop、Skill、Harness、SSE 和完整对话页。
