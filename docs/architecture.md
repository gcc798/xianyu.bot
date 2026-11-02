# 目标架构

## 目标

本项目是服务闲鱼自动化和可审计 Agent 的模块化单体，保留可靠的登录、RBAC、字典、菜单和审计基础能力，删除演示、代码生成、插件市场及无关系统工具。它不是通用低代码平台。

## 目录职责

| 路径 | 职责 | 状态 |
| --- | --- | --- |
| `server/` | Go HTTP 服务和通用后台基础能力 | 主动维护 |
| `server/internal/database/` | PostgreSQL 启动迁移与 Goose 迁移文件 | 主动维护 |
| `server/internal/xianyu/` | 闲鱼、Agent、MCP 和资源业务 | 主动维护 |
| `agent-web/` | React 18 + TypeScript 后台与 Agent 前端 | 主动维护 |
| `script/tel-bot/` | Telegram 资源渠道和 MCP 支撑服务 | 辅助服务 |
| `web/` | Vue 旧后台，只用于迁移对照和回滚 | 已废弃 |
| `.agents/skills/` | 仓库级 Codex 工作流 | 主动维护 |

## 后端边界

`server/internal/xianyu/` 是正式业务模块：

```text
router -> api -> service -> model
                   |
                   +-> protocol / connection / external services
```

- Router 注册路径和中间件。
- API 绑定 HTTP 输入、调用 Service 并返回统一响应。
- Service 负责 owner 校验、状态机、事务和外部调用策略。
- Model 定义持久化结构及请求/响应契约。
- Protocol 和 Connection 隔离闲鱼通信细节。
- Initialize 只负责稳定路由、菜单、权限和种子数据注册，不是插件入口。

不得把闲鱼业务放回 `server/plugin/`，也不得重新设计插件框架。

## 通用后台边界

继续复用：

- 用户登录和 JWT；
- Casbin 角色与 API 权限；
- 用户、角色、菜单、API、字典以及登录/操作/错误日志；
- `{ code, data, msg }` 统一响应；
- 从登录用户获得 owner 身份。

通用后台只承担基础设施，闲鱼业务规则和 Agent 执行必须留在 `server/internal/xianyu/`。

## 前端

`agent-web/` 最终同时承载通用后台、闲鱼业务和 Agent：

- 强制使用 React 18 与 TypeScript。
- 引入后，Ant Design 负责标准后台组件。
- Agent、轨迹、资源和闲鱼业务专属交互可以自研。
- 按能力逐页迁移，并在兼容期保持后端接口不变。
- Vue 只用于确认行为和接口，不作为新架构参考。

## 数据库

- PostgreSQL 是应用数据库。
- `server/internal/database/migrations/` 是生产表结构事实源。
- Goose 在服务启动时执行待应用迁移。
- 复制的 PostgreSQL 结构记录为 `202611010001` 基线。
- 只有系统、闲鱼和 Agent 代表表都存在时，已有数据库才能认领基线。
- 新变化新增前向迁移，基线不可修改。
- GORM `AutoMigrate` 只允许用于测试自己拥有的临时数据库。

## 安全边界

- 用户业务数据在 Service 层校验 `owner_id`。
- Cookie、API Key、MCP 凭据、卡密和 Token 不进入普通响应、日志、浏览器、Prompt、SSE 或轨迹。
- 外部调用前先用短事务持久化意图或资源预占，外部调用不占用数据库事务。
- 远端结果未知时不得当作成功或盲目重试。
- 浏览器不持有 Telegram 等来源凭据。

## 当前迁移状态

后端领域拆分和数据库迁移基础已经建立。React 目前覆盖登录、核心后台和 Agent 对话/轨迹；完整角色授权和多数闲鱼业务页面仍依赖 Vue。

继续阅读[项目范围](project-scope.md)、[当前数据模型](data-model.md)、[功能矩阵](migration/feature-matrix.md)和[前端迁移矩阵](migration/vue-to-react.md)。
