# Vue 到 React 迁移

## 原则

`agent-web/` 是目标前端，`web/` 是冻结的行为证据，不是第二套维护产品。

只有保留行为、鉴权和失败状态在 React 中通过验证后，对应 Vue 页面才能退出。

## 当前 React 路由

| 路由 | 能力 | 状态 |
| --- | --- | --- |
| `#/overview` | 后台总览 | 已实现 |
| `#/agent` | Agent 对话和轨迹 | 已实现 |
| `#/users` | 用户 | 基础 CRUD |
| `#/dictionaries` | 字典和字典项 | 已实现 |
| `#/login-logs` | 登录日志 | 基础列表/删除 |

## Vue 仍承担的保留能力

| Vue 能力 | React 目标 | 状态 |
| --- | --- | --- |
| 个人资料与密码 | 账号设置 | 缺失 |
| 闲鱼总览 | 闲鱼工作台 | 缺失 |
| 账号与扫码登录 | 账号管理 | 缺失 |
| 商品与设置 | 商品管理 | 缺失 |
| 消息与接管 | 消息运营 | 缺失 |
| 回复规则 | 回复自动化 | 缺失 |
| 订单 | 订单管理 | 缺失 |
| 发货规则和任务 | 发货管理 | 缺失 |
| 业务事件 | 业务审计 | 缺失 |
| 模型/MCP/Agent 设置 | Agent 配置 | 缺失 |
| Telegram 运维 | 资源来源管理 | 待定范围并迁移 |

## 组件方向

Ant Design 引入后负责 Layout、Menu、Button、Input、Select、Form、Table、Pagination、Modal、Drawer、Popconfirm、Alert、Message、Notification、Empty、Spin、Tag 和 Badge。

Agent 对话、轨迹、工具检查、资源获取和闲鱼业务工作流交互继续使用产品专属组件。

Ant Design 不负责应用状态管理。先保留 React 局部状态，只有跨页面状态或服务端缓存复杂度明确出现时才增加针对性依赖。

## 单页验收

1. 核对需要保留的 Vue 行为和后端契约。
2. 使用 TypeScript 实现 React 路由。
3. 验证管理员和非管理员权限。
4. 验证加载、空、校验、后端错误和危险操作状态。
5. 验证桌面与窄屏。
6. 执行生产构建和登录态浏览器流程。
7. 记录证据后更新功能矩阵。
8. 在单独可审查的改动中删除 Vue 页面。
