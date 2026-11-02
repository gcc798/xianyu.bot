# 目标架构

## 目标

本项目是服务闲鱼自动化和可审计 Agent 的模块化单体，通用后台只保留登录、固定管理员边界、用户、字典和登录审计。它不是通用低代码平台或通用权限脚手架。

## 目录职责

| 路径 | 职责 | 状态 |
| --- | --- | --- |
| `server/` | Go HTTP 服务和通用后台基础能力 | 主动维护 |
| `server/internal/app/` | HTTP 路由组合根 | 主动维护 |
| `server/internal/database/` | PostgreSQL 启动迁移与 Goose 迁移文件 | 主动维护 |
| `server/internal/api/{router/,}xianyu/` | 闲鱼、Agent、MCP 和资源 HTTP 适配层 | 主动维护 |
| `server/internal/service/xianyu/` | 闲鱼、Agent、MCP 和资源业务服务 | 主动维护 |
| `server/internal/model/xianyu/` | 闲鱼、Agent、MCP 和资源持久化与 API 契约 | 主动维护 |
| `server/internal/xianyu/` | 闲鱼模块装配、通信协议与连接运行时 | 主动维护 |
| `agent-web/` | React 18 + TypeScript 后台与 Agent 前端 | 主动维护 |
| `script/tel-bot/` | Telegram 资源渠道和 MCP 支撑服务 | 辅助服务 |
| `web/` | Vue 旧后台，只用于迁移对照和回滚 | 已废弃 |
| `.agents/skills/` | 仓库级 Codex 工作流 | 主动维护 |

## 后端边界

闲鱼能力是正式业务模块，按 HTTP、业务、模型和通信边界分包：

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
- Initialize 只负责数据库连接、稳定路由和必要的 Agent 默认配置，不是插件入口。
- `cmd/server` 创建进程信号上下文并持有数据库、日志和 HTTP 生命周期；业务 worker 从同一个根上下文派生。
- 数据库通过组合根显式注入，不使用包级数据库或日志变量；日志使用 Zap 的标准全局入口。
- Go module 路径固定为 `github.com/gcc798/xianyu.bot`。

不得把闲鱼业务放回 `server/plugin/`，也不得重新设计插件框架。

## 通用后台边界

继续复用：

- 用户登录和 JWT；
- `authority_id=888` 的管理员边界和普通用户身份；
- 用户、字典以及登录日志；
- `{ code, data, msg }` 统一响应；
- 从登录用户获得 owner 身份。

通用后台只承担基础设施；闲鱼业务规则和 Agent 执行留在 `server/internal/service/xianyu/`，`server/internal/xianyu/` 只负责模块装配、通信协议和连接运行时。

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
- `202611020001` 删除动态权限、演示、生成器、参数、API Token 和通用日志等脚手架表。
- `202611020002` 删除 `sys_users.origin_setting` 旧 Vue 界面配置列。
- 只有系统、闲鱼和 Agent 代表表都存在时，已有数据库才能认领基线。
- 新变化新增前向迁移，基线不可修改。
- GORM `AutoMigrate` 只允许用于测试自己拥有的临时数据库。
- 数据库配置只表达 PostgreSQL，不再保留数据库类型、表名前缀和复数命名等脚手架开关。

## 运行配置

- `server/config.yaml` 和 `server/config.docker.yaml` 使用 `postgres`、`log`、`jwt`、`email` 和 `http` 五个配置段。
- 日志只写入 stdout/stderr，由 Docker 或宿主机负责收集、轮转和保留。
- 配置通过 `-c <path>`、`XIANYU_CONFIG` 或默认 `config.yaml` 选择；启动后不做部分生效的热更新。
- HTTP 路径固定，不提供全局路由前缀开关。
- JWT 撤销记录持久化到 PostgreSQL，并在进程内使用标准库缓存到 Token 过期；不依赖通用缓存服务。

## 安全边界

- 用户业务数据在 Service 层校验 `owner_id`。
- Cookie、API Key、MCP 凭据、卡密和 Token 不进入普通响应、日志、浏览器、Prompt、SSE 或轨迹。
- 外部调用前先用短事务持久化意图或资源预占，外部调用不占用数据库事务。
- 远端结果未知时不得当作成功或盲目重试。
- 浏览器不持有 Telegram 等来源凭据。

## 当前迁移状态

后端已删除动态 RBAC、菜单/API 管理和无关基础设施。React 覆盖最小后台与 Agent 对话/轨迹；多数闲鱼业务页面仍待从 Vue 迁移。

继续阅读[项目范围](project-scope.md)、[当前数据模型](data-model.md)、[功能矩阵](migration/feature-matrix.md)和[前端迁移矩阵](migration/vue-to-react.md)。
