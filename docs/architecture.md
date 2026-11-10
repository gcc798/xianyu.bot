# 架构

## 组成

| 模块 | 职责 |
| --- | --- |
| `server/cmd/server` | 进程入口和依赖组装 |
| `server/internal/api` | HTTP Handler 与 Router |
| `server/internal/service` | 基础能力和闲鱼业务 |
| `server/internal/model` | 持久化模型与 API 契约 |
| `server/internal/agent` | Agent、MCP、Tool、统一资源目录和轨迹 |
| `server/internal/xianyu` | 闲鱼协议、连接和 Tool 适配 |
| `server/internal/infra` | PostgreSQL、日志、邮件和 HTTP 运行设施 |
| `agent-web` | 唯一维护的 Web 前端 |
| `telegram.bot` | Telegram 连接、来源检索和文件下载执行器 |

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
- Telegram 批量扫描与 Agent/MCP 获取共用 Go 资源目录；Telegram 下载完成后由 Go 按用户归属登记统一资源，文件删除也由 Go 同步清理对象和目录记录。
- Telegram 扫描任务和文件运维记录按 Go 代理传入的用户 ID 隔离；资源文件最终记录在 `xy_resources`，来源保存在 `xy_resource_sources`。当前 Telegram 任务的 CRUD 与 worker 进度仍由 `telegram.bot` 持久化，后续可继续迁入 Go；跨进程仅保留排队与取消等执行器通知。
- Telegram API 应用凭据和 FloodWait 截止时间由 Go 管理，Hash 使用现有凭据加密；Telegram Python 服务按配置模式从 Go 内部接口取得最高且未限流的 API ID，任务遇到 FloodWait 后切换配置并继续，连续 5 个不同 API ID 限流时失败。
- Agent 从配置快照读取按优先级排序的模型，使用 Eino 模型 Failover 在同次运行内降级；轨迹只记录模型名称和失败摘要，不记录凭据。
- PostgreSQL 由依赖注入传递，不使用包级数据库。

## Web 前端

`agent-web` 使用 React 19、TypeScript、Vite、React Router 和 Ant Design。应用入口负责全局主题与鉴权上下文，Router 负责页面懒加载和权限守卫，Layout 提供唯一的控制台导航；资源库、采集任务和文件运维归于资源中心，Telegram 菜单只保留连接、会话和设置。

前端只实现当前 Go API 合同，不保留废弃路由、字段或响应结构的运行时兼容逻辑。

## 运行

Docker Compose 启动 Web、Server、PostgreSQL、RustFS 和 Telegram。服务只面向可信内网，日志写入标准输出。
