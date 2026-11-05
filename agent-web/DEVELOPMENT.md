# React 前端开发记录

本文记录 `agent-web/` 的实施约束、当前状态和后续开发批次。`docs/frontend.md` 继续只描述已经落地的前端事实。

具体业务能力、接口、字段、权限和后端前置缺口以 [BUSINESS_CAPABILITY_MIGRATION.md](./BUSINESS_CAPABILITY_MIGRATION.md) 为准。

## 当前基线

- `agent-web/` 是唯一维护和部署的前端。
- 当前工程仍使用 React 18.3；正式页面开发开始前升级到 React 19。
- 当前正式代码包含登录、总览、用户、字典、登录日志以及 Bot 对话、审批和轨迹能力。
- `AdminApp.tsx`、`App.tsx` 和全局 CSS 文件体积过大，正式开发时按页面和职责拆分。
- `web/` 是只读 Vue 参考，只用于确认业务流程、字段、接口和危险操作，不复制其组件体系和视觉皮肤。
- `src/previews/` 与 `ui-preview.html` 是隔离预览，不进入正式路由。

最终状态：React 正式页面全部完成并通过验收后，删除整个 `web/` Vue 工程；仓库只保留 `agent-web/` 作为 Web 前端。

## 已确认技术基线

| 项目 | 选择 |
| --- | --- |
| UI 框架 | Ant Design 6 |
| React | React 19 |
| 语言 | TypeScript |
| 构建 | Vite |
| 路由 | React Router，页面使用懒加载分包 |
| 样式 | CSS Modules + 少量全局 Token |
| 请求 | 保留轻量 `fetch` 封装，统一处理 Token、响应信封和 401 |
| 状态 | 页面状态就地管理；暂不引入 Redux、Zustand |
| UI 依赖 | 不引入 Ant Design 之外的 UI 组件库 |

依赖升级必须同步升级 `react`、`react-dom`、`@types/react` 和 `@types/react-dom`，并执行完整构建与浏览器检查。

## 目录约定

采用 Ant Design Pro 风格的顶层职责分层，同时把页面私有组件放回对应页面目录。

```text
src/
├── app/                 # 应用入口、Provider、主题
├── api/                 # 后端接口，按业务域分文件
├── models/              # DTO、枚举和领域类型
├── router/              # 路由、登录和管理员守卫
├── layouts/             # 全局控制台 Layout
├── components/          # 仅跨页面复用的组件
├── pages/               # 路由页面及页面私有组件
├── hooks/               # 仅跨页面复用的 Hook
├── utils/               # 日期、金额、下载等纯函数
├── styles/              # 全局样式和 Design Token
└── previews/            # 隔离预览，不被正式路由引用
```

约束：

- `pages/` 负责页面组合、查询条件、路由参数和交互状态。
- 页面专用表格、筛选器、Tab、Modal、Drawer 放在页面自己的 `components/`。
- 顶层 `components/` 不接收只服务单个页面的组件。
- `api/` 与 Go、TypeScript 字段保持一致，不使用 `Record<string, unknown>` 代替明确 DTO。
- `models/` 不存放页面临时状态，也不建立通用实体基类。
- 不封装 Ant Design 已有的 Table、Form、Modal、Drawer、Alert、Empty、Skeleton、Tag 和 Pagination。
- 不建立通用 CRUD、万能表单、BaseTable 或推测性的扩展层。

### 新工程约束

- `agent-web/` 按全新 React 工程开发，只实现当前 Go 后端合同，不继承旧 Vue 的技术实现。
- 不编写旧字段别名、旧路由转发、双响应格式解析、降级分支、兼容适配器或迁移期包装层。
- 不使用 `ID ?? id`、`list ?? rows` 等方式同时容纳新旧合同；字段或响应不一致时先统一后端合同和 TypeScript DTO。
- 正式源码中不保留“兼容旧版”“临时过渡”“迁移完成后删除”等兼容性注释、TODO 或废弃代码。
- 中文注释只解释当前业务约束以及代码无法直接表达的原因，不描述旧工程行为和迁移历史。
- `web/` 只用于核对业务能力和接口取舍，不复制其工具函数、请求封装、组件、状态管理或样式。

## 已确认 UI 标准

- 方向为明亮中性、现代企业级运营控制台。
- 所有业务页面使用同一个全局 Layout；Bot 不再表现为另一套产品。
- 侧栏菜单使用“Bot”和“闲鱼自动化”。
- 不展示左下角服务运行状态，不展示无业务价值的页面说明文案。
- 列表页以商品管理预览为基准：紧凑筛选区、40px 数据行、状态胶囊、左下固定分页。
- 详情页使用一级 Tab 划分数据领域；Tab 内采用纵向通栏分区，不使用左右拼图式卡片。
- 字段、表格和分区标题保持统一左右边界，布局横平竖直。
- Bot 输入框固定在会话底部，不展示 Bot 头像；对话与轨迹参考 DeepSeek Harness 的信息结构。
- 当前只验收桌面浏览器；必须处理浏览器缩放后的有效宽度变化，不单独开发移动端页面。
- 保留权限、加载、空、错误、危险操作确认和敏感信息遮罩。

## 页面范围与状态

| 页面 | 正式页面状态 | Vue 参考 | 主要接口 |
| --- | --- | --- | --- |
| 登录与退出 | 已有，待迁移到新结构 | `view/login` | `/base/login`、`/jwt/jsonInBlacklist` |
| 总览 | 已有基础版本，待统一 UI | `xianyu/dashboard.vue` | `/xy/dashboard/overview` |
| Bot | 已有业务逻辑，待按预览重构 | `xianyu/ai-agent.vue` | `/xy/ai/agent/*` |
| 账号管理 | 待开发 | `xianyu/accounts.vue` | `/xy/account/*`、`/xy/qr-login/*`、`/xy/websocket/*` |
| 商品列表 | 预览已确认，正式页待开发 | `xianyu/products.vue` | `/xy/product/list`、`sync`、`setting` |
| 商品详情 | 预览基本确认，正式页待开发 | 后端数据与商品页 | 商品、订单、事件接口 |
| 订单列表与详情 | 详情预览基本确认，正式页待开发 | `xianyu/orders.vue` | `/xy/order/list`、`sync-pending` |
| 消息记录 | 待开发 | `xianyu/messages.vue` | `/xy/message/list`、`reply`、`takeover` |
| 回复规则 | 待开发 | `xianyu/reply-rules.vue` | `/xy/reply-rule/*` |
| 发货规则 | 待开发 | `xianyu/delivery-rules.vue` | `/xy/delivery-rule/*` |
| 发货任务 | 待开发 | `xianyu/delivery-tasks.vue` | `/xy/delivery-rule/task-list`、`task-retry` |
| 业务事件 | 待开发 | `xianyu/business-events.vue` | `/xy/event/list` |
| AI 与 MCP 设置 | 待开发 | `xianyu/ai-settings.vue` | `/xy/ai/models`、`mcp-servers`、`agent-config` |
| Telegram 总览 | 待开发 | `xianyu/telegram-overview.vue` | Telegram 代理接口 |
| Telegram 会话 | 待开发 | `xianyu/telegram-dialogs.vue` | Telegram 代理接口 |
| 资源任务与文件 | 待开发 | `telegram-resource-*.vue` | `/xy/telegram/*`、`/xy/resources/*` |
| Telegram 设置 | 待开发 | `xianyu/telegram-settings.vue` | Telegram 代理接口 |
| 用户、字典、登录日志 | 已有，待拆分和统一 UI | `superAdmin`、`systemTools` | `/user/*`、`/sysDictionary*`、`/sysLoginLog/*` |

旧 Vue 文件只提供参考；正式接口以 `server/internal/api/router/` 和 Go 请求、响应模型为准。

## 开发批次

### 批次 1：工程基础

- 升级 React 19 和对应类型包。
- 安装 React Router。
- 建立 `app`、`api`、`models`、`router`、`layouts`、`components`、`pages`、`styles`。
- 迁移登录、Token、权限守卫和全局 Layout。
- 建立正式 Ant Design Token。

### 批次 2：闲鱼核心链路

- 账号管理与扫码登录。
- 商品列表、商品详情和业务设置。
- 订单列表与订单详情。
- 消息记录、人工回复和接管状态。

### 批次 3：自动化能力

- 回复规则。
- 发货规则、库存导入和发货任务。
- 业务事件。

### 批次 4：Bot 与 AI 配置

- Bot 会话、输入区、审批和错误状态。
- 对话轨迹、工具调用详情。
- 模型、MCP Server 和 Agent 配置。

### 批次 5：Telegram 与系统管理

- Telegram 登录、扫描、会话、资源任务、文件和设置。
- 用户、字典和登录日志迁移。
- 删除被替代的旧 React 大文件。
- React 页面、接口和权限流程全部验收通过后，整体删除 `web/` Vue 工程。

## 最终交付条件

- `agent-web/` 覆盖项目范围内全部正式 Web 页面。
- React 页面使用真实接口完成权限、加载、空、错误和危险操作流程验收。
- `npm run build` 通过，重要业务流程完成 Playwright 浏览器检查。
- Docker、Nginx 和部署入口只构建、发布 `agent-web/`。
- 删除整个 `web/` 目录，并清理所有指向旧 Vue 工程的构建、部署和文档引用。

## 每页完成标准

- 使用真实后端接口，不把模拟数据带入正式页面。
- 列表具备加载、空、错误、筛选、重置和分页状态。
- 写操作具备提交中状态、成功反馈和失败恢复。
- 删除、发货、退出会话、覆盖凭据等危险操作保留二次确认。
- 用户业务数据继续遵守后端 `owner_id` 权限边界；前端不展示无权限入口。
- Cookie、Token、API Key 和凭据不得进入普通页面数据、日志或轨迹。
- 页面路由可直接刷新，浏览器前进后退状态正确。
- 1440px 标准桌面和 1000px 有效宽度下无页面级横向溢出。
- 执行 `npm run build`；重要交互使用 Playwright 检查。

## 开发记录

| 日期 | 内容 | 状态 |
| --- | --- | --- |
| 2026-11-05 | 完成阶段 A 只读审计，确认 `agent-web` 为唯一维护前端 | 完成 |
| 2026-11-05 | 安装 Ant Design，建立隔离 UI 预览入口 | 完成 |
| 2026-11-05 | 确认商品列表、详情信息结构、状态标签和桌面缩放规则 | 完成 |
| 2026-11-05 | 确认采用顶层职责分层，页面私有组件就近存放 | 完成 |
| 2026-11-05 | 确认正式开发使用 React 19 | 已确认，待升级 |

后续每完成一个批次，只追加已经发生的结果、验证命令和剩余问题，不在这里记录逐行修改过程。
