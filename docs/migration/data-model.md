# 新业务数据模型

所有业务主表继承 GVA 时间字段，并显式包含 `owner_id`。外部 ID 使用字符串，金额使用最小货币单位整数；敏感值不参与普通 JSON 序列化。

| 模型 | 关键字段与约束 | 生命周期 |
| --- | --- | --- |
| `xy_accounts` | `owner_id`、`external_user_id`、`note`、`status`、`enabled`；唯一 `(owner_id, external_user_id)` | 禁用保留历史；删除默认软删 |
| `xy_credentials` | `account_id` 唯一、加密 Cookie/Token、状态、过期时间、刷新版本 | 更新采用账号级互斥和版本比较 |
| `xy_products` | `owner_id`、`account_id`、`external_product_id`、标题、价格分、状态；唯一 `(account_id, external_product_id)` | 同步更新，停售不删除 |
| `xy_product_settings` | `product_id` 唯一、回复/发货开关、回复延迟、固定资料 | 随商品保留 |
| `xy_messages` | `owner_id`、`account_id`、会话、外部消息 ID、方向、类型、内容、原始摘要；唯一 `(account_id, external_message_id)` | 可配置保留期；敏感原文不写日志 |
| `xy_conversation_takeovers` | 会话唯一、接管人、开始/结束/过期时间、原因 | 同一会话仅一个活动接管 |
| `xy_reply_rules` | `owner_id`、`product_id`、类型、匹配模式、关键词、优先级、启用 | 删除规则级联内容 |
| `xy_reply_contents` | `rule_id`、文本、媒体引用、权重 | 至少文本或媒体其一非空 |
| `xy_reply_executions` | `message_id` 唯一、策略、状态、响应摘要、错误码 | 防止同消息重复回复 |
| `xy_ai_profiles` | `owner_id`、名称、Base URL、加密 Key、模型、温度、超时、Token 上限 | 响应只返回 Key 是否已配置 |
| `xy_knowledge_documents` | `owner_id`、标题、内容版本、状态、内容哈希 | 更新触发重建 |
| `xy_knowledge_product_links` | 唯一 `(document_id, product_id)` | 保证检索商品隔离 |
| `xy_knowledge_chunks` | `document_id`、序号、文本、内容哈希、向量快照键 | 删除文档同步删除 |
| `xy_vector_index_state` | 快照版本、模型、维度、文件路径、校验和、状态 | 同一 owner 一个活动版本 |
| `xy_orders` | `owner_id`、`account_id`、外部订单 ID、商品、买家、金额分、远端状态；唯一 `(account_id, external_order_id)` | 状态只按允许转换推进 |
| `xy_order_events` | `order_id`、外部事件 ID、类型、载荷摘要；事件 ID 唯一 | 追加写，作为恢复依据 |
| `xy_delivery_rules` | `product_id`、可选 SKU、模式、模板、启用；唯一 `(product_id, sku_key)` | 商品级/SKU 级匹配 |
| `xy_delivery_resources` | 规则/资源池、加密内容、`available/reserved/consumed`、版本 | 原子状态转换，不物理删除已消费资源 |
| `xy_delivery_tasks` | `order_id` 唯一、规则、状态、尝试次数、下次执行、未知结果标记 | 持久化状态机，重启恢复 |
| `xy_delivery_events` | 任务、类型、结果、错误码、时间 | 追加写审计，不存明文卡密 |

## 关系

```text
GVA user (owner)
  -> account -> credential
             -> product -> product_settings
                        -> reply_rule -> reply_content
                        -> delivery_rule -> delivery_resource
                        -> knowledge_document (many-to-many)
             -> message -> reply_execution
             -> order -> order_event
                      -> delivery_task -> delivery_event
```

## 状态机约束

- 账号：`disabled -> checking -> online | expired | verification_required | error`。
- 回复执行：`pending -> running -> sent | skipped | failed`，`message_id` 唯一。
- 发货任务：`pending -> reserving -> delivering -> verifying -> confirmed`；任一步可到 `manual_required`，明确未发送才允许 `retry_wait`。
- 发货资源：`available -> reserved -> consumed`；任务失败且确认未泄露时才可回到 `available`。

## SQLite 策略

- 本地配置为 WAL、`busy_timeout=5000`、`foreign_keys=ON`、单连接，避免个人规模下多连接 PRAGMA 不一致和写锁竞争。
- 外部调用前提交“准备/预占”事务，调用后以第二个短事务记录结果。
- 业务唯一索引承担幂等的最后防线；服务层仍返回可识别的重复事件结果。
- 向量快照不是数据库事务的一部分：先写临时文件并校验，再原子替换，最后更新活动快照元数据。
