# 当前数据模型

本文件只记录当前 `server/internal/xianyu/model/` 已存在的业务表。字段、索引和约束的最终事实以 Go Model 和 Goose 迁移为准。

## 账号与运行状态

| 表 | 职责 |
| --- | --- |
| `xy_accounts` | owner 隔离的闲鱼账号和启用状态 |
| `xy_credentials` | 账号加密凭据、版本和过期状态 |
| `xy_settings` | 闲鱼模块内部设置 |
| `xy_business_events` | 脱敏业务运行事件 |

## 商品、消息与自动化

| 表 | 职责 |
| --- | --- |
| `xy_products` | 闲鱼商品快照和远端状态 |
| `xy_product_settings` | 商品回复、发货等业务开关 |
| `xy_messages` | owner 隔离的收发消息 |
| `xy_conversation_takeovers` | 会话人工接管状态 |
| `xy_reply_rules` | 商品回复规则和优先级 |
| `xy_reply_contents` | 回复内容 |
| `xy_reply_tasks` | 延迟合并与持久化回复任务 |

## 订单、发货与续售

| 表 | 职责 |
| --- | --- |
| `xy_orders` | 订单状态快照 |
| `xy_order_events` | 幂等订单事件 |
| `xy_delivery_rules` | 商品或 SKU 发货规则 |
| `xy_delivery_resources` | 加密发货库存及预占状态 |
| `xy_delivery_tasks` | 可恢复的发货状态机 |
| `xy_relist_tasks` | 售出后的续售任务 |

## Agent 与配置

| 表 | 职责 |
| --- | --- |
| `xy_ai_models` | 模型连接配置 |
| `xy_mcp_servers` | MCP Server 配置 |
| `xy_agent_config` | Agent 运行限制和默认配置 |
| `xy_agent_sessions` | owner 隔离会话 |
| `xy_agent_messages` | 会话消息 |
| `xy_agent_runs` | 单次 Agent 运行 |
| `xy_agent_run_tools` | 完整工具调用记录 |
| `xy_agent_events` | 追加式可恢复轨迹事件 |

## 资源

| 表 | 职责 |
| --- | --- |
| `xy_resource_searches` | 资源搜索意图和状态 |
| `xy_resource_candidates` | 来源候选和检查结果 |
| `xy_resource_tasks` | 异步获取与交付任务 |
| `xy_resources` | 已落库资源元数据 |
| `xy_resource_sources` | 资源来源与溯源信息 |

## 跨表约束原则

- 用户业务数据必须保持 `owner_id` 隔离。
- 外部 ID 使用可稳定保存的字符串，不能假设为本地主键。
- 消息、订单事件、Agent 事件和业务事件需要幂等键或追加写语义。
- 卡密和认证凭据必须加密，普通响应和日志不得返回明文。
- 远端操作先持久化意图或预占，明确成功后再推进最终状态。
- 表结构变化只通过新的 Goose 迁移发布。
