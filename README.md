# Xianyu Assistant

Xianyu Assistant 是面向闲鱼自动化和可审计 Agent 工作流的 Go 模块化单体。通用后台只保留登录、管理员/普通用户、字典和登录审计，不再沿用通用权限脚手架、低代码平台和插件市场方向。

## 当前状态

项目正在迁移中：

- 闲鱼业务服务位于 `server/internal/service/`，Agent 能力位于 `server/internal/agent/`。
- PostgreSQL 表结构和最小基础数据由单一 `init.sql` 初始化。
- `agent-web/` 是目标 React 18 + TypeScript 前端。
- `web/` 是冻结的 Vue 旧前端，只暂时用于迁移对照和回滚。
- React 已覆盖最小后台和 Agent 对话/轨迹；多数闲鱼业务页面仍待迁移。

在[迁移矩阵](docs/migration/feature-matrix.md)中的保留能力全部获得 React 验收前，不得删除 Vue 前端。

## 核心能力

- JWT 登录、用户、固定管理员边界、字典和登录日志；
- 闲鱼账号、加密凭据、扫码登录、商品、消息、回复规则、订单、发货和续售；
- Agent 会话、流式输出、MCP 工具、审批、持久化轨迹和工具检查；
- 资源搜索、候选检查、异步获取、来源追踪和交付元数据；
- PostgreSQL 19 初始化、基础数据和容器化启动。

## 目录

```text
server/
  internal/agent/       Agent、MCP 与资源能力
  internal/api/         HTTP Handler 和 Router
  internal/service/     通用基础与闲鱼业务服务
  internal/model/       持久化实体与请求/响应契约
  internal/infra/       PostgreSQL、日志与 HTTP 基础设施
  internal/xianyu/      闲鱼协议、连接与启动逻辑

agent-web/              主动维护的 React 前端
web/                    已废弃的 Vue 迁移对照
telegram.bot/           Telegram 资源与 MCP 支撑服务
docs/                   架构、领域、范围和迁移状态
.agents/skills/         仓库级 Codex Skills
```

## 技术栈

后端使用 Go、Gin、GORM、PostgreSQL、JWT、Eino 和 MCP。前端使用 React 18、TypeScript、Vite、原生 Fetch 和流式响应。Ant Design 已确定为标准组件方向，但尚未安装。

## 文档入口

- [项目协作规则](AGENTS.md)
- [目标架构](docs/architecture.md)
- [当前数据模型](docs/data-model.md)
- [范围与验收](docs/project-scope.md)
- [闲鱼领域](docs/domains/xianyu.md)
- [Agent 领域](docs/domains/agent.md)
- [Vue 到 React 迁移](docs/migration/vue-to-react.md)
- [PostgreSQL 初始化](docs/database.md)

根目录 `docker-compose.yaml` 是当前唯一保留的容器编排入口：

```bash
docker compose config
docker compose up -d --build
```

首次登录使用 `admin / 123456`，登录后应立即修改密码。PostgreSQL 19 当前仅有官方 RC 镜像，正式版发布后应将 Compose 镜像切换到稳定的 `19-alpine`。

常用验证：

```bash
cd server
go test ./internal/agent/... ./internal/service/... ./internal/xianyu/...

cd ../agent-web
npm run build
```

真实闲鱼账号和交易验收需要明确授权，不能用单元测试结果替代。

## 上游与许可

本仓库基于 Gin-Vue-Admin 二次开发，仍包含继承的后端和 Vue 旧前端代码。根目录 [LICENSE](LICENSE) 为 Apache License 2.0。其他项目只作为协议行为参考，具体边界见[上游与许可说明](docs/upstream-and-licensing.md)。
