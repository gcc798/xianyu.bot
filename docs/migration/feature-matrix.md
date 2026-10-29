# 功能迁移矩阵

状态说明：`可用` 表示旧源码存在完整主流程；`部分` 表示存在明显缺口或只能在特定协议条件下工作；`待实测` 表示依赖真实闲鱼接口，静态审计不能确认。

| 原业务功能 | 原代码位置 | 业务规则与当前状态 | 策略 / GVA 复用 | 新表 | 新接口 | 优先级 | 验收方式 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 后台用户与登录 | `controller/LoginController.java`、`entity/SysUser.java` | 独立 JWT、登录 Token；可用但与 GVA 重复 | **REUSE** GVA 用户/JWT/Casbin | 否 | 否 | P0 | GVA 登录、登出、角色权限回归 |
| 系统菜单与前端 Layout | `vue-code/src/components/layout/`、`router/index.ts` | 独立导航和登录页 | **REUSE** GVA Layout、动态菜单、主题 | 否 | 否 | P0 | 菜单按角色显示、深色模式 |
| 操作日志 | `OperationLogController.java`、`XyOperationLog.java` | 通用与业务日志混合 | **REUSE** GVA 操作日志；业务事件 **REDESIGN** | 业务事件表 | 查询接口 | P2 | 写操作审计且不含凭证 |
| 系统设置 | `SysSettingController.java`、`XySysSetting.java` | 键值表含 AI/邮件等配置 | **REUSE** Viper/SysParams；敏感项 **REDESIGN** | AI 凭证表 | 配置接口 | P1 | 脱敏读写、热加载测试 |
| 数据备份 | `backup/`、`DataBackupController.java` | 按模块 JSON 导入导出 | **REDESIGN** 为 SQLite/向量快照备份 | 否 | 管理接口 | P3 | 备份恢复到临时目录后校验 |
| 闲鱼账号 | `AccountController.java`、`AccountServiceImpl.java` | 备注、UNB、deviceId、状态；部分方法依赖外部 API | **REDESIGN**，owner CRUD/启停与登录态检查已完成 | 账号 | CRUD/启停/检查 | P1 | 已通过两用户数据隔离与状态检查测试 |
| Cookie 管理 | `XyCookie.java`、`CookieRefreshServiceImpl.java` | hasLogin、Set-Cookie 合并、状态更新；待实测 | **ADAPT**；AES-256-GCM 存储、HTTP 检查、Cookie 合并和版本更新已完成 | 账号凭证 | 更新/检查 | P1 | 已验证密文与日志脱敏、同账号刷新互斥和陈旧版本保护 |
| WebSocket Token | `WebSocketTokenServiceImpl.java` | 签名取 Token、重试、验证码/过期分类；待实测 | **ADAPT**；MTOP 签名、Cookie 合并、设备 ID、加密缓存、版本保护与风控分类已完成 | 账号凭证 | 连接内部使用 | P1 | 单元测试与授权账号真实 Token 获取均已通过，响应不记录 Token |
| 二维码登录 | `QRLoginServiceImpl.java`、`QRLoginController.java` | Playwright 会话轮询并提取 Cookie；待实测 | **ADAPT** 为无浏览器 HTTP 会话；创建/轮询/取消/超时清理已完成，人机验证不绕过 | 内存会话 | 创建/状态/取消 | P2 | 授权账号扫码确认、Cookie 加密落库与登录态检查已通过 |
| Cookie 浏览器刷新 | `PlaywrightManager.java`、`CookieRefreshServiceImpl.java` | HTTP 优先、浏览器兜底 | **ADAPT**；Go 方案实测后选型 | 否 | 触发接口 | P2 | Linux/Docker 可用性和 RSS |
| WebSocket 连接 | `WebSocketServiceImpl.java`、`XyWebSocketClient.java` | 注册、心跳、Token 刷新和重连；固定线程池队列无界 | **ADAPT** 协议，**REDESIGN** 生命周期/背压；真实连接、注册、同步状态、ACK、心跳和有界重连已完成 | 进程内连接状态 | 启停/状态 | P1 | 竞态/背压/超时测试及授权账号真实长连接均已通过；页面展示最近错误 |
| 消息解码与路由 | `SyncMessageHandler.java`、`WebSocketMessageRouter.java` | `/s/para`、`/s/sync`，MessagePack/JSON 解析；待实测 | **ADAPT**；解析器、真实传输过滤与持久化路由已完成 | 消息 | 消息查询 | P1 | fixture 回放、空批次、无消息控制推送、已读回执和未知类型测试已通过 |
| 消息去重保存 | `ChatMessageEventSaveListener.java`、schema 唯一索引 | `(account,pnmId)` 去重；可用 | **REDESIGN** owner+外部消息唯一键；已接入 runtime | 消息 | 查询接口 | P1 | 并发重复写入仅保存一条，owner 隔离列表测试通过 |
| 人工接管 | `HumanTakeoverManager.java`、`XyHumanInterventionRecord.java` | 关键词切换、内存超时和记录 | **REDESIGN**；显式持久化接管/释放已完成，人工回复成功后自动接管 | 会话接管 | 接管/释放/回复 | P2 | owner 隔离、接管阻断与出站记录测试通过 |
| 商品同步与详情 | `ItemServiceImpl.java`、`ItemDetailSyncServiceImpl.java` | 列表/详情 API；浏览器详情兜底明确未实现 | **ADAPT**；真实在售商品分页同步、Cookie 更新、幂等落库与下架标记已完成，详情/SKU 待后续 | 商品、SKU | 同步/列表/详情 | P1 | 单元测试通过；授权账号实测同步 8 个在售商品并在页面展示 |
| 商品业务开关 | `XyGoodsConfig.java` | 自动回复、关键词、上下文、发货等聚合 | **REDESIGN**；回复/发货开关与延迟已完成 | 商品设置 | 设置接口 | P1 | owner+账号+商品隔离测试已通过 |
| 关键词回复 | `KeywordReplyServiceImpl.java` | 多关键词模糊规则按优先级处理，未命中时兜底 | **ADAPT**；每商品多条模糊规则、一条兜底、优先级排序和真实 WebSocket 文本发送已完成，暂为单文本 | 回复规则/内容/任务 | CRUD/发送 | P2 | 优先级、历史精确规则兼容、owner 隔离、回执和超时分类已通过 |
| AI 回复与润色 | `ReplyStrategyResolver.java`、`AIReplyStrategy.java` | 关键词+AI 时先匹配再润色，否则 RAG；失败回退 | **REDESIGN** Prompt 边界与注入防护 | AI 配置/执行记录 | 配置/测试 | P2 | 恶意输入、超时、降级测试 |
| 延迟合并回复 | `AutoReplyDelayServiceImpl.java` | 同会话新消息取消旧任务并重计时；仅内存 | **REDESIGN**；256 会话有界调度、持久任务状态、发送前二次校验和重启恢复已完成 | 待回复任务 | 配置/状态 | P2 | 连续消息只发送最后一条、取消及重启恢复测试 |
| RAG 文档与检索 | `AIServiceImpl.java`、`DynamicVectorStoreManager.java` | 商品 ID 过滤、TopK=5、阈值可配；直接保存 JSON | **ADAPT** 业务隔离，**REDESIGN** 持久化 | 文档/分块/索引状态 | 文档/检索/重建 | P2 | 跨商品零泄漏、损坏快照恢复 |
| Embedding 模型切换 | `DynamicVectorStoreManager.java` | 配置变化重建内存对象，但旧向量兼容性未校验 | **REDESIGN** 显式模型版本与全量重建状态 | 索引状态 | 重建接口 | P2 | 维度变化拒绝加载并重建 |
| 订单同步 | `OrderServiceImpl.java`、`PendingOrderPollService.java` | API 拉取、详情补全；依赖非公开接口 | **ADAPT**；消息入库和待发货 API 同步已完成 | 订单/订单事件 | 同步/列表 | P2 | 授权账号当前返回卖家接口无权限，明确报错 |
| 文本自动发货 | `AutoDeliveryServiceImpl.java`、`TextDeliveryStrategy.java` | 付款消息触发，WebSocket 或虚拟发货 API | **REDESIGN**；规则、持久任务与真实虚拟发货 API 已接入 | 发货规则/任务 | CRUD/执行 | P2 | 远端成功才确认；网络未知禁止盲重试 |
| 卡密管理与发货 | `KamiConfigServiceImpl.java`、`KamiDeliveryStrategy.java` | 获取卡密后记录使用；并发/发送失败回收边界不足 | **REDESIGN**；加密库存、预占、确认消耗已完成 | 发货资源/预占 | 导入/库存/执行 | P2 | 两任务争抢仅一条成功；密文落库 |
| 待发货轮询 | `ApiDeliveryScheduler.java`、`PendingOrderPollService.java` | 进程内周期任务拉取并触发 | **REUSE** GVA Timer + 持久任务 | 发货任务 | 状态/重试 | P2 | 重启恢复、重复扫描不重复发货 |
| 手动发货/确认发货 | `AutoDeliveryController.java`、`OrderController.java` | 可手动触发、虚拟发货与确认；待实测 | **ADAPT** 协议，服务端 owner/状态校验 | 发货任务/事件 | 操作接口 | P2 | 非法状态拒绝、结果核验 |
| 图片上传与图片消息 | `ImageUploadServiceImpl.java`、`WebSocketController.java` | 上传本地/URL，发送图片消息 | **REUSE** GVA 上传；协议发送 **ADAPT** | 否 | 引用/发送 | P3 | URL 校验、大小限制、SSRF 测试 |
| 仪表盘与数据面板 | `DashboardController.java`、`DataPanelController.java` | 账号、商品、订单、收入统计 | **REDESIGN** 聚合查询，复用 GVA 页面组件 | 否 | 汇总接口 | P3 | 统计口径与 owner 隔离 |
| 邮件通知 | `EmailNotifyServiceImpl.java` | Cookie 过期、断线、发货失败通知 | **REUSE** GVA email 插件并做业务适配 | 通知偏好可选 | 设置接口 | P3 | 限频、敏感内容检查 |

## 审计结论

- `REUSE` 不是少数：用户、登录、权限、菜单、通用日志、配置框架、任务调度、上传、前端 Layout/主题/请求全部直接复用 GVA。
- 旧项目的真实价值集中在闲鱼协议行为、异常分类和业务触发条件；这些标记为 `ADAPT`，不会复制 Java 分层或表结构。
- 发货、RAG、并发调度和对象级权限需要 `REDESIGN`，因为旧实现不能满足幂等、原子持久化或多后台用户隔离要求。
