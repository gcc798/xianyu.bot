# XyAssistant 重构架构

## 基线与边界

- GVA 基线：`e1c3cc415b5c3638e904924e78898bba10169a92`，Go 1.24，GVA v2 插件注册机制。
- 旧业务参考：`IAMLZY2018/XyAssistant` 的 `f28432b3c2916ba4d37e261f6e26698b5ac72b14`（v2.0.4）。参考仓库只读。
- 协议实现参考：[zhinianboke/xy-auto-reply](https://github.com/zhinianboke/xy-auto-reply)。仅在对接闲鱼登录、Cookie、HTTP/WebSocket、消息、商品、订单、发货等接口时查阅其当前实现，减少对私有协议的无谓试错；它不是本项目的功能范围或架构基线。
- 部署形态：单进程 GVA 后端、GVA 原生 Vue 管理台、SQLite、按需启动的浏览器进程。
- 当前阶段不实现 ProductLayout、微服务、消息队列、Redis 或向量数据库。

使用上述协议参考时，只提取可验证的协议事实，例如请求地址、参数、请求头、签名、ACK、消息结构和错误分类，并在本项目的 `protocol`/`service` 边界内独立实现，以授权账号脱敏样本和回放测试确认。不得直接照搬其 Python/React 代码、数据模型、服务拆分或功能清单；引用实现时记录所查看的分支或提交，避免上游变化造成结论漂移。该参考项目采用 AGPL-3.0，任何代码复用必须先单独评估许可证影响。

## GVA 能力复用清单

| 能力 | 当前实现位置 | 决策 |
| --- | --- | --- |
| 用户、登录、JWT | `server/model/system/sys_user.go`、`server/middleware/jwt.go` | REUSE；删除旧项目平行用户与登录 Token 设计 |
| 角色、菜单、API 权限 | `server/middleware/casbin_rbac.go`、`server/service/system/sys_menu.go` | REUSE；业务对象归属在 Xy Service 中另行校验 |
| 插件发现与注册 | `server/utils/plugin/v2/`、`server/plugin/register.go` | REUSE；Xy 作为 v2 插件注册 |
| 菜单和动态路由 | `SysBaseMenu`、`plugin-tool/utils.RegisterMenus` | REUSE；稳定前缀 `/xy/` |
| API 目录与 Casbin | `SysApi`、`RegisterApis`、Casbin enforcer | REUSE；安装时仅为内置管理员授权 |
| 统一响应、分页 | `model/common/response` | REUSE |
| 操作日志 | `middleware.OperationRecord`、`SysOperationRecord` | REUSE 写操作审计；协议/任务结果另建业务事件记录 |
| 系统配置 | Viper、`SysParams` | REUSE 普通配置；Cookie、Token、API Key 等敏感数据不写普通参数表 |
| 任务调度 | `global.GVA_Timer` / robfig cron | REUSE 周期扫描；关键任务状态仍持久化 |
| SQLite | `initialize/gorm_sqlite.go`、glebarez/sqlite | REUSE；单写连接、WAL、busy timeout、外键约束 |
| 上传文件 | GVA 本地上传与媒体库 | REUSE；业务表只保存引用 |
| 前端布局、主题、请求 | Layout、Element Plus、UnoCSS、`utils/request.js` | REUSE |

## 插件结构

```text
server/plugin/xy/
  initialize/  # SQLite、API、菜单、权限和路由注册
  router/      # 仅绑定路由与中间件
  api/         # HTTP 参数、当前 GVA 用户、统一响应
  service/     # 数据归属、业务状态机、事务与外部适配编排
  model/       # 持久化模型及 request/response
  protocol/    # 闲鱼 HTTP/WebSocket 推送与历史消息协议适配
  runtime/     # 账号级连接、有界工作队列、心跳和重连
  rag/         # 后续：内存精确检索和原子 JSON 快照
web/src/plugin/xy/
  api/
  view/
  components/  # 仅在页面重复使用后再增加
```

依赖方向固定为 `Router -> API -> Service -> Model`。协议 JSON 只存在于 `protocol`，业务层只接收内部事件。Service 接收 `context.Context` 和明确的 `ownerID`，不依赖 `gin.Context`。

## 核心运行流

1. API 从 JWT 获取 GVA 用户 ID。
2. Service 的首个查询条件始终包含 `owner_id`，再解析账号、商品或订单。
3. 账号运行管理器为每个启用账号持有一个可取消生命周期；消息进入有界队列，满载时记录背压错误，不无限创建 goroutine。
4. 协议适配器把外部消息转成内部事件，持久化唯一 `external_message_id` 后再触发回复或订单流程。
5. 自动回复按“人工接管 > 精确关键词 > 模糊关键词 > 商品兜底 > AI/RAG”选择唯一结果。
6. 自动发货通过持久化任务状态机执行；先向买家发消息，再确认平台发货；外部结果未知时进入 `unknown`，禁止直接重发。
7. 自动续售只在发货确认后创建幂等任务；发布前确认原商品已不在售，成功后把商品级设置和规则迁移至新商品。

## API 规划

所有接口使用 `/xy` 前缀、JWT、Casbin 和 GVA 响应格式。

| 模块 | 代表接口 | 说明 |
| --- | --- | --- |
| 基础状态 | `GET /xy/status` | 当前已实现，真实检查数据库连接 |
| 账号 | `/xy/account/*` | 当前已实现 CRUD、启停和加密凭证写入；所有操作校验 owner |
| 商品 | `/xy/product/*` | 已实现真实在售商品同步、业务开关和幂等落库 |
| 消息 | `/xy/message/*` | 已实现消息查询、人工接管与人工回复；实时推送缺失时由 IM 历史请求补偿 |
| 回复 | `/xy/reply-rule/*` | 当前已实现精确/模糊/兜底规则、文本内容、优先级和启停 |
| AI/RAG | `/xy/ai/*`、`/xy/knowledge/*` | 配置测试、文档、重建和检索调试 |
| 发货规则 | `/xy/delivery-rule/*` | 已实现文本/卡密模板、SKU 优先匹配、库存预占和任务状态 |
| 订单/发货 | `/xy/order/*`、`/xy/delivery-rule/task-*` | 已实现订单同步、消息触发、发消息和平台确认发货 |
| 业务日志 | `/xy/events/*` | 按账号、商品、订单和时间过滤 |

## 菜单与页面

一级菜单固定为 `Xy` / `/xy`，当前子页为工作台、账号管理、商品管理、回复规则和发货规则。后续按模块增加消息、自动发货和日志；业务页不读取 Layout 内部状态。

## 安全与资源约束

- 凭证字段不进入普通列表响应、操作日志或错误文本；落盘前采用主密钥加密，主密钥只来自环境或受控配置。
- 本地启动前使用 `export XY_CREDENTIAL_KEY="$(openssl rand -base64 32)"` 设置主密钥；生产环境必须固定并备份该值，丢失后已有凭证无法恢复。
- 不实现验证码绕过、风控规避或“模拟真人”延时。
- SQLite 使用短事务；外部 HTTP/WebSocket 调用不在写事务内。
- 浏览器只为明确任务启动，人工验证出现后释放资源并标记账号待处理。
- 旧项目的 WebSocket 固定线程池使用无界队列，新实现必须使用有界队列和明确并发上限。

## 已知风险

1. 闲鱼协议并非稳定公开契约，正式实现前必须用授权账号抓取脱敏样本并建立回放测试。
2. 旧项目以“消息发送成功”近似发货成功，且卡密可能先占用后发送；新状态机必须区分 `sent`、`confirmed`、`unknown`。
3. 旧向量存储直接覆盖 JSON，缺少模型/维度校验；新实现使用临时文件、校验、替换和快照版本。
4. GVA 插件注册辅助函数不自动为角色授权；当前骨架只显式授权内置管理员 `888`，其他角色由 GVA 权限管理配置。
