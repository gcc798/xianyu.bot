# 架构

## 组成

| 模块 | 职责 |
| --- | --- |
| `server/cmd/server` | 进程入口和依赖组装 |
| `server/internal/api` | HTTP Handler 与 Router |
| `server/internal/service` | 基础能力和闲鱼业务 |
| `server/internal/model` | 持久化模型与 API 契约 |
| `server/internal/agent` | Agent、MCP、Tool、资源和轨迹 |
| `server/internal/xianyu` | 闲鱼协议、连接和 Tool 适配 |
| `server/internal/infra` | PostgreSQL、日志、邮件和 HTTP 运行设施 |
| `agent-web` | 唯一维护的 Web 前端 |
| `telegram.bot` | Telegram 资源渠道 |

## 依赖方向

```text
Router -> API -> Service -> Model
                 -> Xianyu

API -> Agent -> Tool
             -> MCP
             -> Model

Xianyu Tool -> Service
```

- `cmd/server` 负责组装，不承载业务。
- Service 不依赖 Gin，也不持有 Agent。
- Agent 通过 Tool 使用闲鱼业务；进程内调用不走 MCP。
- 外部资源能力通过 MCP 接入。
- PostgreSQL 由依赖注入传递，不使用包级数据库。

## Web 前端

`agent-web` 使用 React 19、TypeScript、Vite、React Router 和 Ant Design。应用入口负责全局主题与鉴权上下文，Router 负责页面懒加载和权限守卫，Layout 提供唯一的控制台导航；API、DTO、页面和跨页面组件按职责分层。

前端只实现当前 Go API 合同，不保留废弃路由、字段或响应结构的运行时兼容逻辑。

## 运行

Docker Compose 启动 Web、Server、PostgreSQL、RustFS 和 Telegram。服务只面向可信内网，日志写入标准输出。
