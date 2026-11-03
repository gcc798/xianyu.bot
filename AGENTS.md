# 仓库规则

## 工程边界

- `server/`：Go 后端。
- `agent-web/`：唯一维护的 React 前端。
- `telegram.bot/`：Telegram 资源服务。
- `web/`：待删除的旧 Vue 参考，不再修改或构建。
- `docs/`：只记录当前事实，不记录迁移过程。

## 后端

- `api/` 处理 HTTP 输入输出，`service/` 处理业务，`model/` 定义数据，`api/router/` 注册路由。
- `service/` 平铺基础能力和闲鱼业务，不恢复 `system`、`xianyu` 子包。
- `agent/` 独立承载 Agent、MCP、Tool 和资源能力。
- `xianyu/` 只承载闲鱼协议、连接和 Agent Tool 适配。
- Service 不依赖 `gin.Context`，Agent 与 Service 不互相持有；Agent 通过 Tool 调用业务能力。
- 直接依赖所需对象，不增加无业务价值的包装、接口或中间层。
- 用户业务数据必须在 Service 校验 `owner_id`。

## Go 风格

- 使用明确的领域名称，避免无实际含义的 `manager`、`processor`、`runtime`、`platform`。
- 导出的类型、接口、方法和函数使用简短简体中文 GoDoc。
- 结构体字段使用简短的中文尾行注释。
- 构造函数命名为 `NewXxx`，放在对应类型附近。
- 修改 Go 代码必须执行 `gofmt` 和受影响测试。

## API 与安全

- 响应保持 `{ code, data, msg }`，分页保持 `{ page, pageSize, total, list }`。
- Go 与 TypeScript 的字段名和类型保持一致。
- API 路径、方法、鉴权或响应变化必须同步调用方。
- Cookie、Token、API Key、卡密和 MCP 凭据不得进入普通响应、日志、SSE、Prompt 或轨迹。
- 当前部署边界是可信内网；公网部署需要另行设计网关、限流和认证。

## 数据库

- PostgreSQL 结构和基础数据只维护 `server/internal/infra/migrations/init.sql`。
- `init.sql` 只用于空库，不在服务启动时建表或执行生产 `AutoMigrate`。
- 测试拥有的临时数据库可以使用 `AutoMigrate`。
- 基础表使用 `s_`，闲鱼表使用 `xy_`，Telegram 表使用 `tel_`。
- 不提交真实账号、凭据、会话或业务数据。

## 前端

- 新功能只写入 `agent-web/`，使用 TypeScript 和 `.tsx`。
- 保留鉴权、权限、加载、空、错误、响应式和危险操作确认。
- UI 修改必须执行 `npm run build`；重要交互还需真实浏览器检查。
- `web/` 只允许读取，待剩余页面取舍完成后整体删除。

## 文档与验证

- 架构变化更新 `docs/architecture.md`，范围变化更新 `docs/project-scope.md`。
- 不复制源码结构、表字段或历史过程到文档。
- 先运行最小相关测试，再运行受影响应用的完整检查。
- 不得把模拟测试写成真实账号或真实交易验收。
- 遵守 `LICENSE` 和 `docs/upstream-and-licensing.md`。

## Skills

- 修改 Go、React、TypeScript 或 Vue 业务代码时使用 `.agents/skills/idiomatic-business-code/`。
- UI 设计、审计和视觉调整时使用 `.agents/skills/impeccable/`。
