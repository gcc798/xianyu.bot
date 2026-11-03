# 目标架构

## 目标

本项目是服务闲鱼自动化和可审计 Agent 的模块化单体，通用后台只保留登录、固定管理员边界、用户、字典和登录审计。它不是通用低代码平台或通用权限脚手架。

## 目录职责

| 路径 | 职责 | 状态 |
| --- | --- | --- |
| `server/` | Go HTTP 服务和通用后台基础能力 | 主动维护 |
| `server/internal/infra/` | PostgreSQL、日志、HTTP 和数据库初始化等运行基础设施 | 主动维护 |
| `server/internal/api/` | 平铺的 HTTP Handler；`router/` 创建 Gin 引擎并注册路径和中间件 | 主动维护 |
| `server/internal/service/` | 平铺的通用基础和闲鱼业务服务 | 主动维护 |
| `server/internal/model/` | 平铺的持久化实体，`request/` 和 `response/` 存放 API 契约 | 主动维护 |
| `server/internal/agent/` | Agent 运行、会话、事件、Tool 契约、MCP 和资源能力 | 主动维护 |
| `server/internal/xianyu/` | 闲鱼通信协议、连接运行时、启动检查和 Agent Tool 适配 | 主动维护 |
| `agent-web/` | React 18 + TypeScript 后台与 Agent 前端 | 主动维护 |
| `telegram.bot/` | Telegram 资源渠道和 MCP 支撑服务 | 辅助服务 |
| `web/` | Vue 旧后台，只用于迁移对照和回滚 | 已废弃 |
| `.agents/skills/` | 仓库级 Codex 工作流 | 主动维护 |

## 后端边界

闲鱼能力是正式业务模块，按 HTTP、业务、模型和通信边界分包：

```text
cmd/server
├── router -> api -> service -> model
│                         └── xianyu protocol / connection
├── api -> agent -> model / MCP / resource
└── xianyu/agenttool -> service + agent/tool
```

- Router 注册路径和中间件。
- API 绑定 HTTP 输入、调用 Service 并返回统一响应。
- Service 负责 owner 校验、状态机、事务和外部调用策略。
- Model 定义持久化结构及请求/响应契约。
- Protocol 和 Connection 隔离闲鱼通信细节。
- `cmd/server` 是进程组合根，分别创建 Service、Agent、Tool、API 和 Router；`service.New` 只组装基础与闲鱼业务 Service。
- Agent API 直接依赖 Agent 组件，不通过 `service.Service` 间接访问。
- Agent 只调用统一 Tool 契约；本进程闲鱼能力由 `internal/xianyu/agenttool` 直接适配 Service，外部能力由 MCP Tool 适配。
- Tool Code 是稳定机器标识，Tool Name 是展示名称；本地工具和 MCP 工具共用预算、审批、超时、审计和轨迹治理。
- Agent 和闲鱼启动检查分别由各自包的 `Initialize` 负责，不是插件入口。
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

通用后台能力与闲鱼业务服务在 `server/internal/service/` 平铺管理；Agent 执行与 MCP/资源能力留在 `server/internal/agent/`。闲鱼 Tool Adapter 是 Agent 与业务 Service 之间的显式桥梁，不使用本进程 MCP，也不把整个 `service.Service` 注入 Agent。

## 前端

`agent-web/` 最终同时承载通用后台、闲鱼业务和 Agent：

- 强制使用 React 18 与 TypeScript。
- 引入后，Ant Design 负责标准后台组件。
- Agent、轨迹、资源和闲鱼业务专属交互可以自研。
- 按能力逐页迁移，并在兼容期保持后端接口不变。
- Vue 只用于确认行为和接口，不作为新架构参考。

## 数据库

- PostgreSQL 是应用数据库。
- `server/internal/infra/migrations/init.sql` 是表结构和最小基础数据的唯一初始化入口。
- PostgreSQL 容器只在空数据目录执行该脚本，Go 服务不再执行运行时迁移。
- 基础数据只包含超级管理员、字典、Agent 默认上限和 Telegram 默认设置。
- 当前基线不接管旧数据库；既有部署升级前必须另行提供增量迁移方案。
- GORM `AutoMigrate` 只允许用于测试自己拥有的临时数据库。
- 数据库配置只表达 PostgreSQL，不再保留数据库类型、表名前缀和复数命名等脚手架开关。

## 运行配置

- `server/config.yaml` 和 `server/config.docker.yaml` 使用 `postgres`、`log`、`jwt`、`email` 和 `http` 五个配置段；数据库密码和 JWT 密钥通过环境变量注入。
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
