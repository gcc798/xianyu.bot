# 后端开发规范

## 1. 适用范围

本规范适用于后端分层、API、鉴权、安全、数据模型和数据库相关任务。涉及 Go 代码时，还必须同时遵守 `.specs/go-code-style.md`。

## 2. 分层与依赖

- `api/` 处理 HTTP 输入输出，`service/` 处理业务，`model/` 定义数据，`api/router/` 注册路由。
- `service/` 平铺基础能力和闲鱼业务，不恢复 `system`、`xianyu` 子包。
- `agent/` 独立承载 Agent、MCP、Tool 和资源能力。
- `xianyu/` 只承载闲鱼协议、连接和 Agent Tool 适配。
- Service 不依赖 `gin.Context`，Agent 与 Service 不互相持有；Agent 通过 Tool 调用业务能力。
- 用户业务数据必须在 Service 校验 `owner_id`。
- 沿用现有目录、分层、命名和依赖注入方式，不得无故引入新架构层、设计模式、通用框架或第二套基础库。
- 稳定且重复出现的业务逻辑应合理复用；未出现实际复用的逻辑不得提前抽象。

## 3. API 与安全

- 响应保持 `{ code, data, msg }`，分页保持 `{ page, pageSize, total, list }`。
- Go 与 TypeScript 的字段名和类型保持一致。
- API 路径、方法、鉴权或响应变化必须同步调用方。
- Cookie、Token、API Key、卡密和 MCP 凭据不得进入普通响应、日志、SSE、Prompt 或轨迹。
- 当前部署边界是可信内网；公网部署需要另行设计网关、限流和认证。

## 4. 数据库

- PostgreSQL 空库结构和基础数据基线维护在 `server/internal/infra/migrations/init.sql`。
- 后续结构或基础数据变更只新增 `server/internal/infra/migrations/goose/` 增量迁移，不得回写初始化基线。
- `init.sql` 只用于空库；已有数据库由 Go 服务启动时按 Goose 版本自动升级。不得在服务启动时执行生产 `AutoMigrate`，也不得对已有库重放 `init.sql`。
- 测试拥有的临时数据库可以使用 `AutoMigrate`。
- 基础表使用 `s_`，闲鱼表使用 `xy_`，Telegram 表使用 `tel_`。
- 不得提交真实账号、凭据、会话或业务数据。
