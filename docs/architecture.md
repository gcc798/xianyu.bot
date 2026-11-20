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
- 闲鱼本地 Tool 通过 Service 提供商品、订单、消息、自动化规则、发货库存与任务、续售任务和业务事件能力；账号配置、扫码登录及 Cookie/Token 等凭据不开放给 Agent。
- 闲鱼写 Tool 先持久化为待审批调用，只有用户批准后才以当前 run 的 owner 执行；用户可拒绝，拒绝会记录轨迹且不调用业务 Service。续售仅允许重试明确失败的任务。
- 卡密值不进入 Agent Tool 参数或轨迹；导入由发货规则页面直接提交给受 owner 校验和加密保护的 Service。
- 外部资源能力通过 MCP 接入。MCP 控制面把服务配置、健康探测和 Tool 目录刷新相互分离：健康探测使用独立短连接且不刷新目录，目录快照持久化后原子发布到内存，服务暂时离线时仍保留最近一次成功目录。
- Agent 路由先在权限、启停和 allow/deny 过滤后选择 MCP，再选择 Tool；默认只向模型提供小型工作集，低置信度或复杂任务扩大候选，运行中的受控 Tool Search 可加载同一已选 MCP 的其他获准工具。路由 Card 和轨迹不包含凭据、完整外部指令或未压缩 Schema。
- MCP 执行面按配置使用共享连接或 Owner 隔离连接，首次建连合并、调用并发执行，可选服务级并发槽；空闲连接按 TTL 回收，不设置全局活跃连接数量上限。健康检查和目录刷新不进入执行连接缓存。
- MCP 写 Tool 提议保存当时的目录和策略摘要，用户批准前重新校验；定义、启停或 allow/deny 已变化时旧审批失效，不能以新能力执行旧参数。
- Telegram 批量采集任务、来源目录、选源策略和任务 worker 由 Go 管理；Go 通过 Telegram MCP 搜索消息，并使用短期对象上传地址调用 Python 下载文件，随后按用户归属登记到 `xy_resources` 和 `xy_resource_sources`。
- Python 仍管理 Telegram Session，并暂时保留会话扫描/退群、成人模式设置和历史文件运维所需的数据库访问；这些业务数据尚未全部迁移到 Go。Python 的旧批量采集 worker 已停止启动，避免与 Go 同时消费 `tel_resource_tasks`。
- Telegram API 应用凭据和 FloodWait 截止时间由 Go 管理，Hash 使用现有凭据加密；Telegram Python 服务按配置模式通过 Compose 自动生成、随专用卷持久保存的服务令牌从 Go 内部接口取得最高且未限流的 API ID，任务遇到 FloodWait 后切换配置并继续，连续 5 个不同 API ID 限流时失败。
- Agent 从配置快照读取按优先级排序的模型，使用 Eino 模型 Failover 在同次运行内降级；轨迹只记录模型名称和失败摘要，不记录凭据。
- PostgreSQL 由依赖注入传递，不使用包级数据库。

## Web 前端

`agent-web` 使用 React 19、TypeScript、Vite、React Router 和 Ant Design。应用入口负责全局主题与鉴权上下文，Router 负责页面懒加载和权限守卫，Layout 提供唯一的控制台导航；资源库、采集任务和文件运维归于资源中心，Telegram 菜单只保留连接、会话和设置。

前端只实现当前 Go API 合同，不保留废弃路由、字段或响应结构的运行时兼容逻辑。

系统设置的 MCP 页面分别呈现服务健康和缓存目录状态。管理员可并发检查全部服务健康，也可单独刷新某个服务目录；健康不可用但仍有缓存目录是有效且可辨识的状态。

## 运行

Docker Compose 启动 Web、Server、PostgreSQL、RustFS 和 Telegram。服务只面向可信内网，日志写入标准输出。

发货在发送买家消息后记录发送时间，再确认平台发货。重试已记录发送时间的任务只执行确认。单实例重启时，发送结果无法确定的任务进入待核对状态。
