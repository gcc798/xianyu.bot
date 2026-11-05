# React 业务能力迁移与 API 对接基线

本文是 `agent-web/` 正式页面开发的业务核对清单。它用于确保旧 Vue 页面中的有效业务能力被准确迁移，并确保 React 请求、TypeScript DTO、权限和交互与当前 Go 后端一致。

本文不记录迁移过程。实施状态发生变化时只更新当前结论；正式接口以 Go 路由、请求模型和响应模型为最终依据。

## 1. 核对依据与冲突处理

按以下优先级确认业务和接口：

1. `server/internal/api/router/`：HTTP 方法、路径、鉴权边界。
2. `server/internal/model/request/`、`server/internal/model/response/`：字段名、类型、分页结构。
3. `server/internal/api/` 与 `server/internal/service/`：校验、权限、状态约束和副作用。
4. `web/src/api/xianyu/`：旧前端实际调用方式，只用于核对业务能力和接口取舍。
5. `web/src/view/xianyu/`：筛选项、操作流程、确认提示和页面取舍，仅作业务参考。
6. `agent-web/src/previews/`：已确认的信息结构和视觉标准，不作为数据合同。

出现冲突时不得照搬 Vue，也不得在 React 内猜测或同时支持多个字段。先以 Go 代码修正文档和 TypeScript DTO；如果后端没有所需能力，则将页面标记为阻塞，补齐后端合同后再开发。

### 1.1 新工程禁止兼容实现

`agent-web/` 是新的 React 工程，不承担旧 Vue 工程的运行时兼容责任：

- 只调用当前 Go Router 注册的正式路径，只解析当前 Go DTO 定义的字段和响应结构。
- 不增加旧路由代理、字段别名、双格式解析、版本判断、降级分支或新旧 API 适配器。
- 不通过 `unknown`、可选字段堆叠或 `A ?? B` 隐藏前后端合同不一致。
- 不复制 Vue 请求封装、组件、Store、工具函数或 CSS；业务流程经核对后按 React 工程重新实现。
- 正式源码不出现兼容性 TODO、废弃代码以及“临时支持旧版”“迁移后删除”等注释。
- 必要中文注释只说明当前业务规则、权限边界或异常原因，不记录迁移过程。

发现接口不一致时，直接修正唯一的正式合同及其调用方；不得用兼容代码绕过。

## 2. 全局接口合同

### 2.1 主 API

- 除 `/base/login` 和 `/health` 外，本文涉及的业务路由均需要 JWT。
- JWT 优先通过 `x-token` 请求头传递，后端也支持同名 Cookie；用户身份只以后端解析出的 Claims 为准，React 不得用 `x-user-id` 等自报头部建立权限。
- 后端续签时通过 `new-token` 和 `new-expires-at` 响应头返回新令牌，请求层必须统一接收并更新登录态。
- 普通 JSON 接口使用统一信封：`{ code, data, msg }`，`code === 0` 表示业务成功。
- 业务失败可能仍使用 HTTP 200，React 请求层必须同时检查 HTTP 状态和 `code`。
- 未认证响应使用 HTTP 401；请求层统一清理本地登录态并跳转登录页。
- 分页请求统一使用从 1 开始的 `page` 和 `pageSize`，后端将 `pageSize` 限制为最大 100。
- 分页响应统一使用 `{ page, pageSize, total, list }`。
- 金额字段由后端以分返回，例如 `priceCent`、`amountCent`；只在展示层格式化，不改 DTO 含义。
- 时间字段有 ISO 时间和毫秒时间戳两种形式，必须按具体 DTO 处理，不做全局猜测转换。
- Go JSON 字段中的 `ID`、`CreatedAt`、`UpdatedAt` 保持原大小写，TypeScript 不自行改成 `id`、`createdAt`。

### 2.2 Telegram 代理例外

`/xy/proxy/telegram/*` 是鉴权后的透明代理。旧 Vue 在浏览器端把 Telegram JSON 包装为 `{ code: 0, data: upstreamResponse }`，这不是 Go 主 API 的响应信封。

React 应在 `api/telegram.ts` 内单独处理：

- JSON 请求直接按 Telegram 上游结构建模，不让页面感知旧 Vue 的人工包装。
- 导出、预览和下载按 `Blob` 处理，不经过 JSON 信封解析。
- 上游错误必须保留 HTTP 状态和安全错误信息，不能无条件改写为成功。
- Telegram 字段使用上游的 snake_case；不得为了统一外观而在页面临时改名。

### 2.3 权限与数据边界

- 管理员固定权限标识为 `authorityId === 888`。
- AI 模型、MCP Server、用户、字典和登录日志管理仅管理员可见、可访问。
- 闲鱼账号、商品、订单、消息、规则和任务由后端 Service 校验 `owner_id`；前端筛选不能替代后端授权。
- 菜单守卫只改善体验，不能被视作安全边界。
- 删除、停止连接、取消登录、退出 Telegram 会话、重试发货、覆盖凭据等有副作用的操作必须保留确认或明确的提交状态。

### 2.4 凭据字段

Cookie、Token、密码、API Key、MCP 认证令牌和卡密只允许出现在对应的受控写入表单或专用授权流程中：

- 列表、详情、普通响应、日志、SSE、Bot 对话和轨迹不得回显原文。
- 编辑凭据时使用“留空表示不替换”的后端语义，不用已有值回填输入框。
- 账号凭据提交字段为 `cookie`、`token`；AI 模型字段为 `apiKey`；MCP 字段为 `authToken`；Telegram 2FA 字段为 `password`。
- 卡密导入使用 `values: string[]`，成功后页面只刷新库存计数，不展示已提交原文。
- 旧 Telegram `/api/admin/settings/ai/key` 明文读取能力不得迁移到普通 React 页面；若该能力仍有业务必要，需先调整为符合仓库安全边界的专用后端合同。

## 3. TypeScript 对接规则

正式代码按业务域建立明确类型，禁止继续使用 `Record<string, unknown>` 代替 DTO：

```text
src/api/auth.ts              src/models/auth.ts
src/api/accounts.ts          src/models/account.ts
src/api/products.ts          src/models/product.ts
src/api/orders.ts            src/models/order.ts
src/api/messages.ts          src/models/message.ts
src/api/automation.ts        src/models/automation.ts
src/api/agent.ts             src/models/agent.ts
src/api/ai-config.ts         src/models/ai-config.ts
src/api/telegram.ts          src/models/telegram.ts
src/api/system.ts            src/models/system.ts
```

每个 API 函数只做四件事：声明方法和路径、传递已建模参数、解析统一信封、返回明确 DTO。页面负责查询状态和交互，不在 API 层保存 UI 状态。

## 4. 页面与接口迁移清单

状态定义：

- **已有**：正式 React 已有业务实现，仍可能需要拆分或统一 UI。
- **待开发**：后端合同已存在，可直接开发。
- **后端前置**：页面设计已存在，但缺少满足页面所需的数据接口。
- **需核验**：透明代理或外部服务合同需要联调确认。

### 4.1 登录与系统管理

| 页面/能力 | 方法与路径 | 请求与响应要点 | 权限 | 状态 |
| --- | --- | --- | --- | --- |
| 登录 | `POST /base/login` | 登录请求与 `UserResponse`、Token 返回以 `api/user.go` 为准 | 公开 | 已完成 |
| 退出 | `POST /jwt/jsonInBlacklist` | 撤销当前 JWT，成功后无论页面缓存如何都清理本地登录态 | 登录用户 | 已完成 |
| 当前用户 | `GET /user/getUserInfo` | 返回当前用户资料和权限标识 | 登录用户 | 已完成 |
| 修改密码 | `POST /user/changePassword` | 成功后按后端策略处理重新登录 | 登录用户 | 已有，待拆分 |
| 个人资料 | `PUT /user/setSelfInfo` | 只提交允许用户自改的字段 | 登录用户 | 已有，待拆分 |
| 用户管理 | `POST /user/admin_register`、`POST /user/getUserList`、`DELETE /user/deleteUser`、`PUT /user/setUserInfo`、`POST /user/resetPassword` | 列表分页；删除、重置密码二次确认 | 管理员 | 已完成 |
| 字典 | `/sysDictionary/*`、`/sysDictionaryDetail/*` | 路径及方法严格使用 router 定义；父字典与明细联动 | 管理员 | 已完成 |
| 登录日志 | `GET /sysLoginLog/getLoginLogList`、`DELETE /sysLoginLog/deleteLoginLog` | 删除需确认 | 管理员 | 已完成 |

### 4.2 总览

Vue 参考：`web/src/view/xianyu/dashboard.vue`。

| 能力 | 方法与路径 | 请求 | 响应 | 状态 |
| --- | --- | --- | --- | --- |
| 账号筛选 | `GET /xy/account/list` | `page=1&pageSize=100` | 账号分页 | 待开发 |
| 运营总览 | `GET /xy/dashboard/overview` | `accountId?`、必填 `startTime`、`endTime` | `DashboardOverview`：`metrics`、消息趋势/来源、订单趋势、自动化汇总、最近事件 | 待开发 |

页面保留今日、7 日、30 日和自定义区间。趋势时间桶使用响应的 `bucketMillis`，不能由前端固定猜测。

`GET /xy/status` 仍是后端基础状态接口，但不对应独立页面，也不在侧栏或页面左下角持续展示“服务运行正常”。

### 4.3 账号管理

Vue 参考：`web/src/view/xianyu/accounts.vue`、`web/src/api/xianyu/account.js`。

| 能力 | 方法与路径 | 请求/响应 | 交互要求 | 状态 |
| --- | --- | --- | --- | --- |
| 列表 | `GET /xy/account/list` | `page`、`pageSize`、`keyword`、`status`、`enabled?` | 加载、空、错误、分页 | 已完成 |
| 新建 | `POST /xy/account/create` | `{ label, note }` | 成功后刷新 | 已完成 |
| 编辑/启停 | `PUT /xy/account/update` | `{ ID, label, note, enabled }` | 开关失败时恢复原值 | 已完成 |
| 删除 | `DELETE /xy/account/delete` | `{ ID }` | 二次确认 | 已完成 |
| 保存凭据 | `PUT /xy/account/credential` | `{ ID, cookie, token }` | 只写不回填；提交后可重新检查登录 | 已完成 |
| 检查登录 | `POST /xy/account/check-login` | `{ ID }`；响应 `LoginStatus` | 展示状态，不展示凭据 | 已完成 |
| 开始扫码 | `POST /xy/qr-login/start` | `{ ID }`；响应 `QRLogin` | 打开二维码并开始轮询 | 已完成 |
| 查询扫码 | `GET /xy/qr-login/status` | `sessionId` | 成功、失败、过期时停止轮询 | 已完成 |
| 取消扫码 | `DELETE /xy/qr-login/cancel` | `sessionId` | 关闭弹窗和卸载页面时清理 | 已完成 |
| 连接状态 | `GET /xy/websocket/status` | `{ ID }`；响应 `WebSocketStatus` | 按账号刷新 | 已完成 |
| 启停连接 | `POST /xy/websocket/start`、`POST /xy/websocket/stop` | `{ ID }` | 停止连接需确认 | 已完成 |

### 4.4 商品列表与商品详情

Vue 参考：`web/src/view/xianyu/products.vue`。已确认的正式 UI 以隔离预览中的商品列表和详情信息结构为准。

| 能力 | 方法与路径 | 请求/响应 | 状态 |
| --- | --- | --- | --- |
| 商品列表 | `GET /xy/product/list` | `ProductSearch` → `PageResult<Product>` | 已完成 |
| 同步商品 | `POST /xy/product/sync` | `{ accountId }` | 已完成 |
| 商品业务设置 | `PUT /xy/product/setting` | `UpdateProductSetting` | 已完成 |
| 商品详情 | 当前无详情路由 | 不能获得稳定的单商品详情合同 | 后端前置 |
| 关联订单 | 当前只有 `GET /xy/order/list`，不支持商品 ID 筛选 | 不得在前端拉取全量订单后拼接 | 后端前置 |
| 商品最近事件 | `GET /xy/event/list?productId=...` | `BusinessEventSearch` → 分页事件 | 待开发 |

商品列表字段以 `response.Product` 为准：

- `ID`、`CreatedAt`、`UpdatedAt`
- `accountId`、`accountLabel`
- `externalProductId`、`title`、`priceCent`、`status`
- `imageUrl`、`imageUrls`
- `virtualProduct`、`autoReplyEnabled`、`autoDeliveryEnabled`、`autoRelistEnabled`
- `replyDelaySeconds`、`relistDelaySeconds`、`relistDescription`、`relistImageUrls`、`relistAddress`

后端明确不返回 `SourceJSON`。正式商品详情页不得使用预览模拟订单、模拟事件或列表缓存冒充详情数据。后端应先提供按本地 `ID` 获取商品详情的接口，并为关联订单提供明确筛选合同；接口补齐后再更新本文和 DTO。

### 4.5 订单列表与订单详情

Vue 参考：`web/src/view/xianyu/orders.vue`。

| 能力 | 方法与路径 | 请求/响应 | 状态 |
| --- | --- | --- | --- |
| 订单列表 | `GET /xy/order/list` | `page`、`pageSize`、`keyword`、`accountId`、`status` → `PageResult<Order>` | 已完成 |
| 同步待发货 | `POST /xy/order/sync-pending` | `{ accountId }` | 已完成 |
| 订单详情 | 当前无详情路由 | 不能获得稳定的单订单详情合同 | 后端前置 |
| 订单最近事件 | `GET /xy/event/list?orderId=...` | 分页事件 | 业务事件页已完成；详情页后端前置 |

订单列表字段为 `ID`、`CreatedAt`、`UpdatedAt`、`accountId`、`accountLabel`、`externalOrderId`、`externalProductId`、`productTitle`、`buyerUserId`、`buyerUserName`、`sessionId`、`skuKey`、`status`、`amountCent`。

正式订单详情页必须等待按本地 `ID` 查询的后端合同；不能仅凭列表行数据宣称详情迁移完成。

### 4.6 消息记录

Vue 参考：`web/src/view/xianyu/messages.vue`。

| 能力 | 方法与路径 | 请求/响应 | 交互要求 | 状态 |
| --- | --- | --- | --- | --- |
| 消息列表 | `GET /xy/message/list` | `page`、`pageSize`、`accountId`、`sessionId`、`direction`、`externalProductId`、`buyerName`、`content`、`startTime`、`endTime` → `PageResult<Message>` | 时间筛选使用毫秒时间戳 | 已完成 |
| 人工回复 | `POST /xy/message/reply` | `{ messageId, text }` | 禁止空文本；防重复提交 | 已完成 |
| 人工接管 | `PUT /xy/message/takeover` | `{ messageId, active }` | 失败时恢复开关 | 已完成 |

消息 DTO 保留 `contentType`、`direction`、`source` 和 `takeoverActive`，页面不能只按文本消息假设所有内容类型。

### 4.7 回复规则

Vue 参考：`web/src/view/xianyu/reply-rules.vue`。

| 能力 | 方法与路径 | 请求/响应 | 业务约束 | 状态 |
| --- | --- | --- | --- | --- |
| 列表 | `GET /xy/reply-rule/list` | `page`、`pageSize`、`keyword`、`productId`、`type` | 加载商品选项 | 已完成 |
| 新建/编辑/优先级 | `POST /xy/reply-rule/save` | `{ ID, productId, type, keyword, priority, enabled, content }` | 类型为精确、模糊或兜底；兜底唯一性以 Service 校验为准 | 已完成 |
| 删除 | `DELETE /xy/reply-rule/delete` | `{ ID }` | 二次确认 | 已完成 |

### 4.8 发货规则、库存与任务

Vue 参考：`delivery-rules.vue`、`delivery-tasks.vue`。

| 能力 | 方法与路径 | 请求/响应 | 交互要求 | 状态 |
| --- | --- | --- | --- | --- |
| 规则列表 | `GET /xy/delivery-rule/list` | `page`、`pageSize`、`keyword`、`productId`、`mode` → `PageResult<DeliveryRule>` | 文本/卡密模式 | 已完成 |
| 保存规则 | `POST /xy/delivery-rule/save` | `{ ID, productId, skuKey, skuName, mode, template, enabled }` | 提交中防重复 | 已完成 |
| 删除规则 | `DELETE /xy/delivery-rule/delete` | `{ ID }` | 二次确认 | 已完成 |
| 导入卡密 | `POST /xy/delivery-rule/resource-import` | `{ ruleId, values: string[] }` | 原文只提交，不回显 | 已完成 |
| 库存 | `GET /xy/delivery-rule/inventory` | `{ ID }` → `{ ruleId, available, reserved, consumed }` | 展示汇总计数 | 已完成 |
| 任务列表 | `GET /xy/delivery-rule/task-list` | `page`、`pageSize`、`keyword`、`accountId`、`status` → `PageResult<DeliveryTask>` | 状态、重试次数、下次执行、错误 | 已完成 |
| 重试任务 | `POST /xy/delivery-rule/task-retry` | `{ ID }` | 仅失败任务显示；确认后提交 | 已完成 |

### 4.9 业务事件

Vue 参考：`web/src/view/xianyu/business-events.vue`。

`GET /xy/event/list` 接收 `page`、`pageSize`、`keyword`、`accountId`、`productId`、`orderId`、`type`、`status`，返回 `PageResult<BusinessEvent>`。字段包括账号、商品、订单、事件类型、结果状态、摘要和安全错误描述。业务事件是只读审计页，不提供前端伪造、重放或编辑能力。

状态：已完成。

### 4.10 Bot 会话、轨迹与审批

Vue 的 `ai-agent.vue` 只是入口占位；正式业务主要参考现有 `agent-web` 实现和后端合同。

| 能力 | 方法与路径 | 请求/响应 | 状态 |
| --- | --- | --- | --- |
| 会话列表/新建 | `GET /xy/ai/agent/sessions`、`POST /xy/ai/agent/sessions` | 分页；新建 `{ title }` | 已完成 |
| 会话详情/删除 | `GET /xy/ai/agent/sessions/:id`、`DELETE .../:id` | `AgentSessionDetail` | 已完成 |
| 发送消息 | `POST .../:id/messages` | `{ message, idempotencyKey, directives }`；directive 最多 2 个 | 已完成 |
| 取消运行 | `DELETE .../:id/runs/:runId` | 仅运行中允许 | 已完成 |
| 审批运行 | `POST /xy/ai/agent/runs/:id/approve` | 写工具确认 | 已完成 |
| SSE | `GET .../:id/events` | 断线重连和会话切换清理 | 已完成 |
| 轨迹 | `GET .../:id/trajectory` | `afterId`、`limit<=500`、`runId?`、`type?` | 已完成 |
| 工具详情 | `GET .../:id/tool-calls/:callId` | `AgentToolCallDetail` | 已完成 |

轨迹与工具详情只消费后端已经处理过的安全参数和结果。前端不从其他请求缓存补写轨迹内容，也不把认证材料拼入对话或审批说明。

### 4.11 AI 与 MCP 配置

Vue 参考：`web/src/view/xianyu/ai-settings.vue`。以下路由全部仅管理员可用。

| 能力 | 方法与路径 | 主要字段 | 状态 |
| --- | --- | --- | --- |
| 模型列表/新建 | `GET /xy/ai/models`、`POST /xy/ai/models` | `name`、`provider`、`baseUrl`、`apiKey`、`model`、能力开关、上下文与超时 | 已完成 |
| 模型编辑/删除/默认 | `PUT /xy/ai/models/:id`、`DELETE .../:id`、`PUT .../:id/default` | 编辑时 `apiKey` 留空表示不替换 | 已完成 |
| MCP 列表/新建 | `GET /xy/ai/mcp-servers`、`POST /xy/ai/mcp-servers` | `name`、`code`、`baseUrl`、`authHeader`、`authToken`、超时 | 已完成 |
| MCP 编辑/删除 | `PUT /xy/ai/mcp-servers/:id`、`DELETE .../:id` | 编辑时 `authToken` 留空表示不替换 | 已完成 |
| Agent 配置 | `GET /xy/ai/agent-config`、`PUT /xy/ai/agent-config` | 轮次、工具调用和令牌边界以 Go 请求模型为准 | 已完成 |

删除模型或 MCP Server 必须确认；默认模型、启用状态和支持能力使用后端返回值，不在前端推导。

### 4.12 Telegram 管理

Vue 参考：`telegram-overview.vue`、`telegram-dialogs.vue`、`telegram-settings.vue`、`telegram-resource-tasks.vue`、`telegram-resource-files.vue`。所有路径实际通过 `/xy/proxy/telegram` 转发，当前标记为需核验，开发时必须与运行中的 `telegram.bot` 联调。

| 页面/能力 | 上游方法与路径 | 关键请求 | 状态 |
| --- | --- | --- | --- |
| 状态 | `GET /api/admin/status` | 无 | 需核验 |
| 扫码登录 | `POST /api/admin/login/qr`、`GET /api/admin/login/status` | 轮询生命周期 | 需核验 |
| 2FA | `POST /api/admin/login/2fa` | `{ password }` | 需核验 |
| 扫描 | `POST /api/admin/scan`、`GET /api/admin/scan/status` | 启动与轮询 | 需核验 |
| 会话列表 | `GET /api/admin/dialogs` | 上游分页/筛选 | 需核验 |
| 退出会话 | `POST /api/admin/dialogs/leave` | `{ dialog_ids, confirm: '退出' }`，必须二次确认 | 需核验 |
| 导出会话 | `GET /api/admin/dialogs/export` | Blob 下载 | 需核验 |
| 设置 | `GET /api/admin/settings` | 不接入明文凭据读取 | 需核验 |
| 成人内容开关 | `POST /api/admin/settings/adult-mode` | `{ adult_mode }` | 需核验 |
| AI 设置/探测 | `POST /api/admin/settings/ai`、`POST /api/admin/settings/ai/probe` | `base_url`、`api_key`、`model` | 需核验 |
| 资源选项 | `GET /api/resources/options` | 无 | 需核验 |
| 资源任务 | `POST/GET /api/resources/tasks` | 创建参数、分页以服务合同为准 | 需核验 |
| 取消/重试任务 | `POST /api/resources/tasks/:id/cancel`、`POST .../:id/retry` | 路径 ID | 需核验 |
| 资源文件 | `GET /api/resources/files` | 分页与筛选 | 需核验 |
| 预览/下载 | `GET /api/resources/files/:id/:action` | `action` 仅允许服务支持值；Blob | 需核验 |

### 4.13 统一资源交付

这些接口供 Bot 资源任务和正式资源页使用，与 Telegram 管理代理分开建模：

| 能力 | 方法与路径 | 状态 |
| --- | --- | --- |
| 资源列表 | `GET /xy/resources` | 待开发 |
| 资源详情 | `GET /xy/resources/:id` | 待开发 |
| 下载 | `GET /xy/resources/:id/download` | 待开发 |
| 任务详情 | `GET /xy/resource-tasks/:id` | 已有部分能力，待统一 |
| 取消任务 | `POST /xy/resource-tasks/:id/cancel` | 已有部分能力，待统一 |

`/xy/internal/resource-tasks/*` 是服务间内部回调，不得由 React 调用。

## 5. 页面状态与交互验收

每个正式页面必须覆盖：

- 首次加载骨架或明确加载状态。
- 无数据空状态；筛选后无结果与系统无数据文案可区分。
- 网络错误、业务 `code !== 0`、401 和提交失败。
- 查询、重置、分页、浏览器前进后退和直接刷新。
- 写操作提交中禁用、防重复提交、成功反馈和失败恢复。
- 删除、停止、取消、退出、覆盖配置等危险操作确认。
- 管理员页面的菜单、路由守卫和后端 403/401 兜底。
- 1440px 和约 1000px 有效桌面宽度；浏览器缩放后不出现页面级横向溢出。
- 正式页面不引用 `src/previews/`，不连接模拟数据。

## 6. 后端前置缺口

React 正式开发前或对应页面开发时必须解决：

1. **商品详情接口缺失**：新增按本地商品 `ID` 获取详情的明确路由和响应 DTO。
2. **商品关联订单合同缺失**：订单列表当前不支持 `productId` 或 `externalProductId` 筛选；需确定后端查询方式。
3. **订单详情接口缺失**：新增按本地订单 `ID` 获取详情的明确路由和响应 DTO。
4. **Telegram 合同不是仓库内 Go DTO**：需对运行中的代理逐项联调，固化 React TypeScript 类型及错误结构。
5. **Telegram 明文 Key 读取与仓库安全边界冲突**：React 不迁移该读取动作，后端需决定删除或替换合同。

缺口未解决时可以先完成列表页，但不得用模拟数据、全量列表拼接或 `unknown` 类型绕过。

## 7. 迁移完成判定

单页只有同时满足以下条件才算迁移完成：

1. Vue 页面涉及的有效查询、写操作、确认和异常流程已逐项核对。
2. React 使用明确 TypeScript DTO，字段与 Go JSON 标签一致。
3. 真实接口完成加载、空、错误、权限和危险操作验证。
4. `npm run build` 通过，关键流程使用 Playwright 在真实浏览器检查。
5. 正式路由不依赖 Vue 文件、预览组件或模拟数据。
6. 本文对应行更新为已完成，并记录仍存在的真实限制。

只有全部范围内页面通过上述判定，且 Docker、Nginx、部署和文档均不再引用旧工程后，才能整体删除 `web/`。删除必须一次性执行，不保留半套 Vue 运行入口。
