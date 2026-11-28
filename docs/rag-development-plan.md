# RAG 知识库与智能客服 Agent 开发规划

## 进度概览

| 项目 | 当前状态 |
| --- | --- |
| 当前开发阶段 | Phase 01：基础架构与领域模型设计（进行中） |
| 已完成阶段 | 无 |
| 总体开发进度 | 0 / 11（0%） |
| 当前阻塞问题 | EcomRetrieval 未声明明确许可；闲鱼业务模拟集及自建文档金标准尚待领域人员人工核验 |
| 下一步开发任务 | 领域人员核验闲鱼 Qrels 和文档预期值；向 EcomRetrieval 维护者确认许可，无法确认则替换为许可清晰的数据集；完成后关闭 Phase 01 |
| 最后更新时间 | 2026-11-28 |

> 本文档是 RAG 知识库与智能客服 Agent 后续开发的唯一实施和进度基准。只有任务完成且通过对应验收后，才能勾选任务和阶段。实施中如调整设计，必须同步更新“决策与调整记录”、当前进度、遗留问题和下一步任务。

## 1. 规划依据与现状基线

本规划基于 2026-11-27 的仓库实现，重点复用以下现有能力：

| 现有能力 | 当前实现 | 本次复用方式 |
| --- | --- | --- |
| 后端与数据库 | Go 1.26.10、Gin、GORM、PostgreSQL 19、Goose 增量迁移 | 继续使用现有 API、Service、Model 和迁移体系，不引入第二套 ORM 或业务数据库 |
| Agent | Eino 0.9.22、模型优先级与 Failover、Agent Harness | 复用模型调用、执行循环和时间、步数、模型、Tool、Token 等限制 |
| Tool 与 MCP | 本地 Tool Registry、MCP 目录、路由、连接复用、健康检查 | RAG 以本地只读能力接入 Agent；实时业务数据继续由闲鱼 Tool 提供 |
| 审批与轨迹 | 写 Tool 审批、Agent Run、Tool Call、Agent Event、SSE 回放 | 客服 Agent 复用审批与轨迹机制，扩展检索证据、配置快照和耗时事件 |
| 会话可靠性 | 幂等键、同一 Agent 会话仅允许一个运行、取消与历史恢复 | 复用运行治理思想，新增面向闲鱼会话的串行消费与重复回复防护 |
| 闲鱼消息 | 入站消息幂等落库、自动回复任务、延迟发送、新消息合并、人工接管、未知发送状态 | 客服 Agent 接入现有消息处理链路，不创建第二套闲鱼连接和发信实现 |
| 文件存储 | RustFS、MinIO Go Client、短期预签名 URL、资源归属校验 | 原始知识文档和构建中间产物复用 RustFS，数据库只保存元数据与对象引用 |
| 前端 | React 19、TypeScript、Vite、Ant Design | 在唯一维护的 `agent-web/` 中增加知识库、构建任务和检索实验室页面 |
| 部署 | Docker Compose 管理 Web、Server、PostgreSQL、RustFS、Telegram | 增加 Worker、RabbitMQ；Milvus 作为可选 Compose Profile，不影响默认 pgvector 部署 |

当前需要补齐的关键能力：

- 后端只有 `cmd/server` 进程，MCP 后台任务和 Telegram 资源任务也在 Server 内运行，尚无通用 Worker 进程。
- 尚未引入生产用 RabbitMQ、Transactional Outbox、pgvector、Milvus、Embedding Provider 或 Reranker；已建立独立的本地 Embedding 与检索指标验证工具，但尚未接入业务运行链路。
- Eino 已提供 Document、Embedding、Indexer 和 Retriever 基础接口，但仓库只使用了 Chat Model；现有依赖不能直接完成完整知识构建与 Advanced RAG。
- 当前自动回复是确定性规则匹配，尚无客服 Agent 的意图识别、知识检索、可信度降级和 Agentic RAG。
- 当前 PostgreSQL 镜像是官方 `postgres:19-alpine`，不包含 pgvector 扩展，必须先验证兼容的构建和升级方式。

## 2. 建设目标与功能边界

### 2.1 建设目标

1. 建立可管理、可版本化、可异步构建的知识库，覆盖平台、闲鱼账号和商品三个作用域。
2. 建立与闲鱼业务解耦的 RAG 引擎，统一支持解析、切片、Embedding、检索、融合、重排和上下文构建。
3. 默认使用 pgvector，同时提供可通过配置切换的 Milvus 实现及完整重建、迁移和回退能力。
4. 在现有 Agent Harness 上形成独立的客服 Agent，支持知识问答、实时 Tool、人工接管和敏感操作审批。
5. 建立可复现的离线评估、检索实验室、运行轨迹和可靠性验证体系，使效果与性能结论都有真实数据依据。

### 2.2 功能范围

- 知识来源：商品信息、FAQ、手工文本和常用文档。
- 知识管理：知识库、文档、版本、作用域、商品多对多绑定、切片预览、构建进度、失败重试和重新索引。
- RAG：Metadata Filtering、Dense Retrieval、Keyword Retrieval、Hybrid Search、RRF、Query Rewrite、Reranker、Parent-Child Retrieval 和 Context Builder。
- Agent：运营 Agent 与客服 Agent 隔离，支持意图识别、多轮对话、RAG 问答、Tool Calling、Agentic RAG、人工接管和降级。
- 评估：数据集、实验配置、检索与生成指标、消融实验、Benchmark 和结果对比。
- 运维：异步任务、Outbox、重试、死信、恢复、指标、日志、轨迹和安全审计。

### 2.3 暂不纳入

- 除 pgvector 和 Milvus 之外的向量数据库实现。
- 通用工作流设计器、插件市场或面向任意业务的知识中台。
- 网络爬虫、音视频知识抽取、图片 OCR 和自动生成训练数据；后续由真实需求驱动扩展。
- 直接公网部署、多租户 SaaS 计费和跨地域容灾。
- 用向量库保存订单、库存、物流、账号状态等实时数据；这些数据继续通过受控 Tool 查询。
- 在没有真实 Benchmark 前宣称某个模型、切片策略、向量库或参数更优。

## 3. 设计原则

- 复用现有 `service`、`agent`、Tool、MCP、审批、轨迹、RustFS、Goose 和 owner 校验能力。
- RAG 引擎只依赖通用文档、查询、过滤条件和 Provider 接口，不依赖闲鱼商品、账号或消息模型。
- 业务 Service 负责把平台、账号、商品权限转换为 RAG 的通用检索范围，模型不能自行决定权限过滤条件。
- PostgreSQL 是知识元数据、任务状态、版本指针、Outbox 和权限关系的事实来源；RabbitMQ 只负责投递，不作为任务状态来源。
- 构建版本不可变，只有完整构建并验证成功后才能原子发布；失败版本不得影响当前在线版本。
- Broker 消息只传任务和版本标识，不传文档正文、凭据或完整检索内容。
- 默认选择最少的新组件；只有真实评估证明收益后才启用 Reranker、复杂 Query Rewrite 或 Milvus。
- 所有 Provider 都需要统一契约测试，业务代码不得按具体向量库、Embedding 或解析器写分支。

## 4. 整体技术架构

```text
                         ┌──────────────── agent-web ────────────────┐
                         │ 知识管理 / 构建进度 / 检索实验室 / 轨迹 │
                         └───────────────────┬───────────────────────┘
                                             │ HTTP / SSE
┌──────────────┐  闲鱼消息  ┌────────────────▼────────────────────────────┐
│ Xianyu WS    ├───────────►│ server：API、权限、知识管理、运营 Agent、 │
└──────────────┘            │ 闲鱼连接、人工接管、最终消息发送           │
                            └──────┬──────────────┬──────────────────────┘
                                   │ 事务写入     │ RustFS 预签名访问
                                   ▼              ▼
                            ┌────────────┐   ┌────────────┐
                            │ PostgreSQL │   │   RustFS   │
                            │ 元数据/任务│   │ 原始文档   │
                            │ Outbox/轨迹│   │ 构建产物   │
                            └──────┬─────┘   └────────────┘
                                   │ Outbox Relay
                                   ▼
                            ┌────────────┐
                            │ RabbitMQ   │
                            │ 主队列/重试│
                            │ 死信队列   │
                            └──────┬─────┘
                                   │ 多实例竞争消费
                                   ▼
                   ┌──────────────────────────────────┐
                   │ worker：Outbox、知识构建 Pipeline│
                   │ 解析→清洗→切片→Embedding→索引   │
                   │ 客服推理/重建/迁移/评估/任务恢复 │
                   └───────────────┬──────────────────┘
                                   │ VectorStore
                      ┌────────────┴────────────┐
                      ▼                         ▼
               ┌────────────┐          ┌────────────────┐
               │ pgvector   │ 默认     │ Milvus 可选    │
               │ + PG 关键词│          │ + PG 关键词    │
               └────────────┘          └────────────────┘
```

### 4.1 在线检索链路

```text
业务检索范围解析
  → Query Rewrite（按策略可选）
  → Dense / Keyword 并行召回
  → Metadata 与权限过滤
  → RRF 融合
  → Reranker（按策略可选）
  → Parent-Child 扩展
  → 去重、预算裁剪、Context Builder
  → 带证据回答或 Agent Tool 结果
```

Keyword Retrieval 统一由 PostgreSQL 承担：英文、编号和结构化关键词使用全文索引，中文内容先采用 `pg_trgm` 召回并通过 Benchmark 判断是否需要额外中文分词扩展。Milvus 只替换 Dense Retrieval，不替换知识元数据、权限和关键词事实来源。

## 5. 技术选型及理由

| 领域 | 选择 | 理由与约束 |
| --- | --- | --- |
| 开发语言与框架 | 继续使用 Go、Gin、GORM | 与现有后端一致，领域模型、权限和部署可以直接复用 |
| Agent 编排 | 继续使用 Eino 与现有 Harness | 已具备模型 Failover、Tool、审批、轨迹和预算控制；不引入第二套 Agent 框架 |
| 异步消息 | RabbitMQ，Go 客户端使用 `amqp091-go` | 满足已确定技术决策；支持持久队列、Publisher Confirm、手动 Ack、DLX 和多消费者 |
| 任务一致性 | PostgreSQL 任务表 + Transactional Outbox | 在同一事务内创建业务任务和 Outbox，避免数据库成功但消息丢失；数据库保持最终状态 |
| 默认向量库 | pgvector | 复用 PostgreSQL，默认部署最轻；元数据过滤、事务和权限联查直接，适合当前规模起步 |
| 可选向量库 | Milvus Standalone | 为更大向量规模和独立索引能力保留实现；仅通过 `VectorStore` 接口接入 |
| pgvector 索引 | 默认 HNSW，距离度量跟随 Embedding 模型 | 支持持续增量构建且无需训练集；具体参数必须由 Benchmark 决定 |
| 关键词检索 | PostgreSQL 全文检索 + `pg_trgm` | 不新增 Elasticsearch/OpenSearch；兼顾英文、编号和中文片段召回 |
| 原始文档 | RustFS，继续使用 MinIO Go Client | 已有部署、预签名 URL、对象校验和 owner 隔离能力 |
| RAG 数据结构 | 复用 Eino `schema.Document` 和基础组件接口，在其上定义项目所需 Provider 契约 | 与现有 Eino 生态兼容，同时不把业务绑定到某个 Eino 扩展实现 |
| Embedding | `EmbeddingProvider` 注册机制；首个实现采用 OpenAI-compatible API | 可复用兼容端点生态；模型、维度和批量限制必须进入索引版本快照 |
| Reranker | `Reranker` 可插拔接口，默认关闭 | 先取得无 Reranker 基线，再依据效果和延迟数据决定模型与启用范围 |
| 文档解析 | Parser Registry；文本类内置解析，PDF/Office Provider 在 Phase 01 验证后锁定 | Eino 当前仅内置通用接口和文本解析；避免在未验证质量与许可前固定实现 |
| 前端 | React、TypeScript、Ant Design | 与 `agent-web` 当前技术栈一致 |
| 迁移 | 只新增 Goose 增量迁移 | 符合现有数据库升级基线，禁止回写 `init.sql` 或生产 AutoMigrate |
| 可观测性 | 现有 Zap/Agent Event 为基础，增加 Prometheus 指标和可选 OTLP Trace | 先保持本地可运行，再按部署环境连接 OpenTelemetry Collector |
| 测试 | 单元测试使用现有 Testify/SQLite；Provider、迁移、RabbitMQ、pgvector、Milvus 使用真实组件集成测试 | SQLite 不支持向量、全文和队列语义，关键结论不能由 Mock 替代 |

新增依赖的具体版本在对应阶段实现时锁定并更新第三方许可文档，不在规划阶段写入未经验证的版本号。

## 6. 待决策事项与决策门

| 决策项 | 当前状态 | 决策门与输出 |
| --- | --- | --- |
| PostgreSQL 19 + pgvector 镜像 | 基础兼容性已验证 | 已验证 PostgreSQL 19.0、pgvector 0.8.7、pg_trgm 1.6、HNSW 和过滤查询；迁移、备份恢复在 Phase 03/11 随正式 Schema 继续验证 |
| 默认 Embedding 模型与维度 | Ollama 接入和 1024 维已验证，默认模型待正式 Benchmark | `qwen3-embedding:0.6b` 与 `bge-m3` 已完成连通性、维度、中文语义探针、批量、并发和异常验证；模拟业务集尚待人工核验，公开集尚受许可和下载条件阻塞 |
| PDF/Office 解析实现 | 首个实现锁定 Tika Full Sidecar | Apache Tika 3.3.0 Full 加固定中文 OCR 模型通过 DOCX、XLSX、PPTX、多栏中文 PDF 和扫描中文 PDF 共 5 类金标准；Phase 04 实现 Provider，并用脱敏真实文档持续验证泛化，不把 5 个样本外推为生产准确率 |
| Reranker Provider | 待评估 | Phase 06 在固定召回集上比较收益、P95 延迟与成本；无显著收益则保持关闭 |
| 中文关键词方案 | 初始使用 `pg_trgm` | Phase 06 用 Benchmark 决定是否需要中文分词扩展，不提前增加搜索集群 |
| Milvus 部署版本与拓扑 | Standalone 基础兼容性已验证 | Milvus 3.0.5 官方 Compose 与 Go SDK v3 已通过建库、索引、写入、持久化、加载及 owner/版本过滤检索；Phase 03/11 继续锁定正式 Profile、备份恢复和运维参数 |
| 客服自动发送阈值 | 待业务标注 | Phase 08 基于真实评估集确定；未达到阈值时只生成建议或转人工，不自动回复 |

### 6.1 Phase 01 执行记录

| 日期 | 验证项 | 结果 | 证据与后续 |
| --- | --- | --- | --- |
| 2026-11-27 | RAG 契约与业务解耦 | 通过 | 新增独立 `internal/rag` 契约，复用 Eino Parser、Transformer、Embedder，只为双向量库和高级检索定义必要接口；包测试与 `go vet` 通过 |
| 2026-11-27 | 权限与版本不变量 | 通过 | 强制 owner、授权索引版本、账号/商品关系和向量维度一致；已发布状态由业务 Service 解析后传入；普通 Metadata Filter 不能覆盖授权字段 |
| 2026-11-27 | 索引与任务生命周期 | 通过 | 以状态机覆盖 staging 发布/失败、published 退役、retired 回滚、任务领取、租约恢复、重试和终态 |
| 2026-11-27 | PostgreSQL 19 + pgvector | 通过基础兼容性验证 | 使用 `pgvector/pgvector:0.8.7-pg19-trixie` 临时容器实测 vector/pg_trgm 扩展、HNSW、owner/版本过滤、向量排序和中文 trigram 查询；未接触项目数据库 |
| 2026-11-27 | Milvus Standalone | 通过基础兼容性验证 | 使用官方 v3.0.5 Compose 启动 Milvus、etcd 和对象存储，健康检查通过；使用官方 Go SDK v3 实测建 Collection、向量索引、写入、Flush、Load 与 `owner_id + index_version` 标量过滤，结果未混入其他 owner；验证后已删除临时容器、网络、卷和探针文件 |
| 2026-11-27 | Embedding/Reranker | 受阻 | 当前执行环境无可用模型凭据，且仓库没有代表性标注语料；不得用 Mock 代替质量、吞吐和成本结论 |
| 2026-11-27 | PDF/Office Parser | 通过基础兼容性验证 | Apache Tika 3.3.0 临时容器成功提取合成中文 DOCX/PDF 的标题、订单号、规则文本、简单表格及英文标记；镜像约 171 MB，空闲容器约 244 MiB。缺少真实复杂版面、扫描件和异常文档，暂不锁定最终实现；临时容器已删除，脱敏探针保留在被 Git 忽略的 `.codex-work/` |
| 2026-11-27 | Go 后端验证 | 通过 | 修正既有测试对 `atomic.Int64` 的直接复制，改为读取原子值；`go test ./...`、`go vet ./...` 与差异检查均通过 |
| 2026-11-28 | Ollama Embedding | 通过运行能力验证 | OpenAI Compatible 端点实测两个模型均为 1024 维；批量 32、并发 2、4 次请求共 128 条均成功，无效模型和取消请求均被正确处理。小型语义探针与吞吐数据见 `docs/rag-evaluation-report.md`，不据此判定模型优劣 |
| 2026-11-28 | 检索指标与评估数据 | 工具链通过，质量结论待人工核验 | 建立版本化 Corpus/Queries/Qrels 结构及 Recall@K、Precision@K、MRR、NDCG 实现；闲鱼模拟集扩充为 24 篇知识、32 个查询和 45 条分级 Qrel，输出逐查询 Top 5 证据。两个本地模型完成实测，但数据仍为 `pending_human_review`，不用于最终选型。EcomRetrieval 导入脚本固定上游提交，许可未明确 |
| 2026-11-28 | 复杂 Office 与异常文档 | 通过扩展兼容性验证 | 固定 Apache POI 提交并按需下载 12 个 DOC/DOCX/XLS/XLSX/PPT/PPTX 样本；Tika 处理 11 个，`chartex.docx` 返回 422。空、损坏和加密文件均返回 422，错误后缀按内容提取；无金标准，结构保真和阅读顺序仍未验收 |
| 2026-11-28 | 复杂文档金标准与中文 OCR | 技术选型通过 | 按需生成 DOCX、XLSX、PPTX、多栏中文 PDF 和扫描中文 PDF 及机器可读金标准；Tika Full 加固定 `chi_sim` 模型按文本、顺序、表格和 OCR 规则 5/5 通过，容器约 325.8 MiB。锁定为 Phase 04 首个 Parser Provider；真实文档泛化和人工确认仍是后续验收项 |

参考的一手资料：[pgvector 官方仓库](https://github.com/pgvector/pgvector)、[Milvus Standalone Docker Compose](https://milvus.io/docs/install_standalone-docker-compose.md)、[Apache Tika 官方文档](https://tika.apache.org/)。具体依赖仍以实际实现时锁定并验证的版本为准。

## 7. 模块职责与进程划分

### 7.1 代码模块

| 模块 | 主要职责 | 明确不负责 |
| --- | --- | --- |
| `internal/rag` | 通用文档、解析、切片、Embedding、VectorStore、检索、融合、重排、上下文和评估接口 | 闲鱼账号、商品、消息权限和 HTTP |
| `internal/model` | 知识库、文档版本、构建任务、Outbox、检索配置、实验与结果的持久化模型 | 执行 Pipeline 或权限编排 |
| `internal/service` | 知识库 CRUD、作用域与 owner 校验、商品绑定、任务创建、版本发布、实验管理 | 直接依赖 Gin 或具体向量库 |
| `internal/agent` | 运营/客服 Agent Profile、RAG Tool、轨迹事件和 Harness 集成 | 直接访问闲鱼协议或绕过 Service |
| `internal/xianyu` | 闲鱼连接及 Agent Tool 适配 | 知识构建和通用 RAG 算法 |
| `internal/api` | 输入校验、响应、SSE 和调用 Service | 执行文档构建任务 |
| `cmd/server` | API、现有运营 Agent、闲鱼连接、人工接管、客服回复 ready 事件消费与最终消息发送 | 知识构建 Pipeline、客服推理和 Outbox 后台投递 |
| `cmd/worker` | Outbox Relay、知识构建、客服 Agent 推理、索引重建/迁移、评估任务、恢复与清理 | 对外 HTTP API、闲鱼长连接和最终消息发送 |

`internal/rag` 内允许为 Provider 建立必要的实现子包，但不增加只有转发作用的 Repository、Manager 或通用工作流层。

### 7.2 进程与基础设施

- Server 和 Worker 使用同一代码仓库、领域模型、配置加载和数据库连接设施，构建为独立入口，可使用同一镜像的不同启动命令。
- Worker 支持多实例竞争消费；每条文档构建任务只有一个有效租约，过期任务可以被其他实例恢复。
- Outbox Relay 运行在 Worker 中，多实例通过数据库行锁领取记录，发送成功必须取得 RabbitMQ Publisher Confirm 后才能标记完成。
- RabbitMQ 使用持久 Exchange、主队列、分级 TTL 重试队列和死信队列，不依赖延迟消息插件。
- Milvus 仅在选择该 Provider 时部署；默认 Compose 不启动 Milvus。

### 7.3 Provider 注册

需要注册和配置的扩展点：

- `VectorStore`：pgvector、Milvus。
- `EmbeddingProvider`：首个 OpenAI-compatible 实现，后续按需增加。
- `DocumentParser`：文本、Markdown、HTML、CSV/JSON、PDF/Office Provider。
- `Chunker`：固定窗口、递归、结构化、Parent-Child。
- `QueryRewriter`：关闭、规则、LLM。
- `Reranker`：关闭和模型 Provider。

注册表只根据配置创建实现；任务和检索配置保存 Provider、模型、维度、算法参数及版本快照，运行中不得读取会漂移的隐式默认值。

## 8. 领域与数据设计

### 8.1 领域对象

规划使用 `rag_` 表前缀，与现有 `agent_`、`mcp_`、`resource_` 和 `xy_` 领域区分。具体字段和 SQL 在 Phase 01 设计评审后确定，不回写初始化基线。

- 知识库：名称、作用域、状态、当前已发布索引版本和默认检索策略。
- 知识来源：FAQ、文本、商品快照或 RustFS 文档引用。
- 文档与文档版本：保留来源身份和不可变内容版本；相同内容通过内容摘要幂等识别。
- 索引版本：绑定 Embedding、Chunking、VectorStore 和检索配置快照，记录 staging、published、retired 等生命周期。
- Chunk：保存父子关系、顺序、文本、来源定位、权限元数据和向量记录引用。
- 商品绑定：知识库与商品多对多关系；绑定和解绑必须校验 owner 与商品归属。
- 构建任务：文档级任务、尝试次数、租约、阶段、进度、失败详情和产物版本。
- Outbox：业务事件、聚合标识、消息版本、投递状态和重试信息。
- 检索配置：召回、TopK、阈值、融合、重排和上下文预算的版本化 Profile。
- 评估对象：数据集、样例、实验、运行、指标和证据快照。

遵守仓库数据库规则：不新增物理外键、`CHECK`、数据库 ENUM、DOMAIN 或触发器；关系完整性和业务状态由 Service 校验，必要的唯一约束和索引用于幂等及性能。

### 8.2 作用域与权限

- 平台级：仅管理员维护；普通用户只能检索已发布且允许全平台使用的知识。
- 账号级：必须同时匹配当前 owner 和闲鱼账号。
- 商品级：通过商品绑定解析，必须同时匹配 owner、账号和商品。
- 向量权限元数据使用 `owner_id = 0` 明确表示平台知识；非平台知识禁止使用零 owner，避免把共享内容伪装成某个用户的数据。
- 检索范围由 Service 根据认证用户或入站闲鱼消息构造为不可变 `RetrievalScope`，不得接收模型生成的 owner、账号或商品过滤条件。
- pgvector 查询使用关系条件约束候选；Milvus 查询使用同等 scalar filter，并在返回 PostgreSQL 元数据时再次校验。
- 平台、账号、商品知识可同时召回，优先级和去重规则由检索 Profile 明确配置。

### 8.3 版本与发布

1. 创建文档版本和 staging 索引版本。
2. Worker 将全部 Chunk 写入版本隔离的物理命名空间。
3. 完成数量、维度、来源、权限元数据和抽样检索校验。
4. 在 PostgreSQL 事务中原子切换知识库的 published 指针。
5. 在线查询只读取 published 版本；旧版本进入保留期后异步清理。

Embedding 模型、维度、归一化方式或 Chunking 变化必须产生新索引版本，禁止在原版本上混写不兼容向量。

## 9. 异步任务与可靠性设计

### 9.1 Transactional Outbox

知识管理 Service 在同一 PostgreSQL 事务中创建文档版本、构建任务和 Outbox 记录。Outbox Relay：

1. 使用短事务和 `FOR UPDATE SKIP LOCKED` 分批领取待投递记录。
2. 发布只含任务 ID、文档版本 ID、事件类型和消息 Schema 版本的消息。
3. 等待 Publisher Confirm 后标记已投递；不确定结果保持可重试。
4. 使用 Outbox ID 作为消息 ID，Consumer 通过任务版本和阶段幂等消费。

### 9.2 RabbitMQ 拓扑

- 主 Exchange 按任务类型路由文档构建、重建、迁移、客服推理、回复 ready 和评估任务。
- 主队列使用 durable、manual ack 和合理 prefetch。
- 短、中、长三档 TTL 重试队列通过 DLX 回到主队列，避免 Worker 内阻塞等待。
- 超过最大尝试次数或确认不可重试的任务进入死信队列，同时把数据库任务置为失败。
- 管理端“重试”创建新的尝试或重新入队事件，不直接篡改 Broker 队列。

### 9.3 Worker 幂等、恢复与并发

- 文档版本 + Pipeline 配置具有唯一幂等身份，重复消息不能产生重复 published 版本。
- Worker 领取任务后写入租约和心跳；进程退出后由扫描器回收过期租约。
- 每个 Pipeline 阶段保存可观测进度；可安全重试的阶段允许重入，不可重用的部分先清理 staging 产物。
- 并发限制至少覆盖全局 Worker、单知识库、Embedding Provider 和 VectorStore 写入。
- Ack 发生在数据库状态和阶段产物安全落定之后；数据库更新失败时消息重新投递。

## 10. 核心业务流程

### 10.1 文档上传与知识构建

1. 用户创建知识库或添加 FAQ、文本、商品快照、文档。
2. API 校验 owner 和作用域，通过预签名地址把文件写入 RustFS。
3. Service 校验对象元数据，在事务中创建文档版本、构建任务和 Outbox。
4. Worker 消费任务，按 Parser Registry 完成解析、清洗和结构恢复。
5. Chunker 生成叶子 Chunk 和可选父 Chunk，补齐来源与权限元数据。
6. Embedding Provider 分批生成向量，VectorStore 写入 staging 索引版本；关键词索引写入 PostgreSQL。
7. Worker 执行完整性与抽样检索校验，成功后原子发布，失败则保留当前线上版本并记录可重试错误。

### 10.2 知识更新与重建

1. 修改内容时创建新的文档和索引版本，不直接覆盖 published 数据。
2. 未变更文档可按内容摘要复用解析或 Chunk 产物，但不得复用不兼容的 Embedding。
3. 全部构建成功后切换 published 指针。
4. 回滚只切换到仍在保留期且验证通过的旧版本。

### 10.3 RAG 查询

1. 调用方提交查询和由 Service 生成的 `RetrievalScope`。
2. Engine 加载指定 Profile，执行可选 Query Rewrite。
3. Dense 与 Keyword Retriever 并行召回，应用版本、作用域和 Metadata Filter。
4. 使用 RRF 融合、去重，可选 Reranker 和 Parent-Child 扩展。
5. Context Builder 按 Token 预算组织证据，保留知识库、文档、版本和位置引用。
6. 返回上下文、分数、证据引用、耗时和配置快照；无足够证据时返回低可信状态，不强行生成答案。

### 10.4 智能客服 Agent

1. 复用现有闲鱼 WebSocket 和消息幂等入库，在同一事务中以触发消息 ID 创建客服处理任务和 Outbox。
2. 先检查人工接管、商品开关、消息类型、会话状态和重复任务。
3. Worker 按账号 + 闲鱼会话串行消费客服任务；新消息可以取消尚未发送的旧回复候选。
4. 意图识别后，静态商品/FAQ 知识走 RAG；订单、库存、发货等实时信息走只读 Tool。
5. 简单知识问答走固定 RAG Workflow；需要多步检索或实时业务组合时进入 Agentic RAG。
6. Worker 生成前校验证据、敏感内容和可信度；通过策略后持久化回复候选和 reply-ready Outbox，不直接连接闲鱼发送。
7. Server 消费 reply-ready 事件，复用现有回复任务，在发送前再次检查人工接管、消息时序和重复内容，再通过 `SendText` 发送。
8. 不满足自动发送策略时保存建议、转人工或生成安全降级结果；写操作必须进入现有审批机制，客服 Agent 默认工具白名单只开放只读能力。

运营 Agent 与客服 Agent 共享模型、Harness、Tool 实现和轨迹基础设施，但使用独立的 Prompt、上下文、运行配置、工具白名单和会话数据。客服会话不得出现在运营 Bot 会话列表中。

### 10.5 向量库切换与迁移

1. 创建指向目标 Provider 的新索引版本。
2. 从 PostgreSQL 的已发布 Chunk 与版本元数据重新生成或复用兼容向量，写入目标 Provider。
3. 运行数量校验、权限过滤测试、固定查询集对比和抽样回读。
4. 先灰度只读对比，再原子切换 active Provider/版本。
5. 保留原 Provider 作为回退，达到保留期和验收条件后再清理。

### 10.6 评估实验

1. 选择不可变评估数据集版本和检索 Profile。
2. 运行一个或多个 Embedding、Chunking、Retriever、Reranker、VectorStore 组合。
3. 保存配置、代码版本、数据版本、随机种子、原始排名、证据和耗时。
4. 计算 Recall@K、Precision@K、MRR、NDCG；生成评估计算 Faithfulness，并保留 Judge 模型和 Prompt 版本。
5. 对比实验和消融实验只基于同一数据集版本，结果可导出和复跑。

## 11. 智能客服安全与防护

- 把知识文档和 Tool 返回的外部文本视为不可信数据，不能作为系统指令执行。
- 检索上下文必须以证据区块注入 Prompt，并明确禁止其中内容改变 Agent 权限和工具策略。
- 任何日志、轨迹、RabbitMQ 消息和普通 API 响应不得包含 Cookie、Token、API Key、卡密或完整 MCP 凭据。
- 检索轨迹默认保存证据 ID、来源、分数和安全摘要；完整原文通过 owner 校验后按需读取。
- 人工接管在任务创建、Agent 执行前和最终发送前至少三次检查，接管后不得自动发送。
- 触发消息 ID 保证一次处理；回复内容摘要与时间窗口防止重复发送；发送结果不确定时进入现有 unknown/待核对语义。
- 客服 Agent 设置更保守的执行时间、迭代、模型、Tool 和 Token 上限；达到上限时转人工，不继续循环。
- 平台知识变更、批量绑定、重建和 Provider 切换只允许管理员或明确授权用户执行，并记录审计。

## 12. 评估与可观测性

### 12.1 离线评估

- 检索数据集必须包含查询、相关文档/Chunk、作用域和标注来源。
- 检索指标：Recall@K、Precision@K、MRR、NDCG、无结果率和权限错误率。
- 生成指标：Faithfulness、答案相关性、引用覆盖率、拒答正确率；LLM Judge 必须配合人工抽样校验。
- 工程指标：构建吞吐、失败率、P50/P95 延迟、Embedding 调用量、Token、存储量和成本。
- 所有效果结论附带数据集版本、样本量、配置和运行时间，不用单个演示查询代替 Benchmark。

### 12.2 在线轨迹

- 扩展现有 Agent Event，增加 rewrite、retrieval、fusion、rerank、context 和 answer 等阶段事件。
- 独立记录 Retrieval Run，使检索实验室和非 Agent 调用也能审计；Agent Event 只引用 Retrieval Run 和安全摘要。
- 记录模型、Embedding、Reranker、VectorStore、索引版本、TopK、过滤范围、候选分数、最终证据、Token、耗时和错误码。
- 构建任务记录每阶段开始、完成、数量、耗时、失败类别和重试次数。

### 12.3 指标与告警

- RabbitMQ：积压、最老消息年龄、重试、死信、消费速率和未确认数。
- Worker：活跃租约、过期租约、阶段耗时、成功率和 Provider 限流。
- 检索：QPS、P95、无结果率、低可信率、各 Retriever 候选数和 Reranker 降级率。
- 客服：任务延迟、自动发送率、转人工率、重复拦截、未知发送状态和预算超限。
- 基础设施：PostgreSQL 慢查询、pgvector/Milvus 健康、RustFS 错误和模型服务可用率。

告警阈值在真实压测和灰度数据后确定，不在缺少基线时填写任意数字。

## 13. 分阶段开发任务与验收标准

### Phase 01：基础架构与领域模型设计

**阶段目标**：锁定领域边界、数据生命周期、Provider 契约和关键技术决策，为后续实现提供稳定基础。

**前置依赖**：无。

**核心开发任务**：

- [x] 梳理知识库、文档版本、索引版本、构建任务、Outbox、作用域、检索 Profile 和评估对象的领域关系。
- [x] 定义 RAG Engine、VectorStore、Embedding、Parser、Chunker、Reranker 的行为契约与错误语义。
- [x] 设计 `rag_` 表族、状态迁移、幂等键、租约、版本发布和清理策略。
- [x] 完成 PostgreSQL 19 + pgvector 与 Milvus 部署基础技术验证。
- [x] 完成 Ollama OpenAI Compatible Embedding 的连通性、维度、批量、并发和异常技术验证。
- [ ] 完成许可清晰的公开检索集基线。
- [ ] 完成闲鱼业务标注的领域人员核验。
- [x] 完成 PDF/Office Parser 首个实现技术选型。
- [x] 明确客服 Agent 与运营 Agent 的会话、权限、Prompt、工具和运行限制边界。
- [x] 更新受影响的架构、数据库、项目范围和第三方许可文档设计要求。

**验收标准**：

- 领域评审覆盖正常、重复、失败、重试、取消、恢复、回滚和删除流程。
- Provider 契约不包含闲鱼业务模型，pgvector 与 Milvus 能使用同一契约测试计划。
- 所有待决策项都有实测记录和明确结论；未决事项不得进入依赖它的实现阶段。
- 数据库设计符合 Goose 增量迁移和无物理外键等仓库规则。

**完成状态**：

- [ ] Phase 01 完成

### Phase 02：多进程架构、消息队列及异步任务机制

**阶段目标**：建立独立 Worker、RabbitMQ、Transactional Outbox 和可恢复任务骨架。

**前置依赖**：Phase 01 完成。

**核心开发任务**：

- [ ] 增加 Worker 进程入口、共享配置、优雅退出、健康检查和 Compose 部署。
- [ ] 接入 RabbitMQ，建立持久主队列、分级重试队列、死信队列和手动 Ack。
- [ ] 实现任务状态、Outbox Relay、Publisher Confirm、幂等消费、租约心跳和过期恢复。
- [ ] 增加多 Worker 并发、Provider 级限流、任务取消和管理端重试能力。
- [ ] 建立 RabbitMQ 与 PostgreSQL 真实组件集成测试。

**验收标准**：

- 数据库事务提交后即使 RabbitMQ 短时不可用，恢复后消息仍能投递；事务回滚不会产生消息。
- 重复投递不会重复创建业务产物；杀死 Worker 后过期任务可由另一实例恢复。
- 可重试错误按策略重试，不可重试或超限任务进入死信并在 PostgreSQL 中可查询。
- API 进程不执行知识构建 Pipeline。

**完成状态**：

- [ ] Phase 02 完成

### Phase 03：向量数据库抽象及 pgvector、Milvus 实现

**阶段目标**：提供业务无感的双 VectorStore 实现和可验证的切换能力。

**前置依赖**：Phase 01、Phase 02 完成。

**核心开发任务**：

- [ ] 实现 VectorStore Provider Registry 和共享契约测试。
- [ ] 实现 pgvector 写入、删除、Dense Search、Metadata Filter、健康检查和 HNSW 管理。
- [ ] 实现 Milvus 对等能力及可选 Compose Profile。
- [ ] 建立索引命名、Embedding 维度、版本隔离、批量写入和部分失败语义。
- [ ] 实现重建、双读对比、切换、回退和旧索引清理流程。

**验收标准**：

- 两个 Provider 通过同一套增删查、过滤、隔离、幂等和错误语义测试。
- 固定数据集在两个 Provider 上均能返回正确作用域内结果，不能检索其他 owner 的数据。
- Provider 切换期间在线查询始终使用完整 published 版本，失败可回退。
- pgvector 默认部署可独立运行，不要求启动 Milvus。

**完成状态**：

- [ ] Phase 03 完成

### Phase 04：知识库管理与异步构建 Pipeline

**阶段目标**：完成知识库生命周期、文档管理和文档级异步构建闭环。

**前置依赖**：Phase 02、Phase 03 完成。

**核心开发任务**：

- [ ] 实现平台、账号、商品作用域和知识库与商品多对多绑定。
- [ ] 实现 FAQ、文本、商品快照和 RustFS 文档来源管理。
- [ ] 实现 Parser、Cleaner、Chunker、Embedding、Index Writer Pipeline 和阶段进度。
- [ ] 实现文档版本、索引版本、切片预览、失败重试、重建、原子发布和旧版本清理。
- [ ] 为解析、切片、幂等、权限和版本切换建立测试语料与回归测试。

**验收标准**：

- 支持的每类来源都能完成上传、构建、预览、发布、更新、失败重试和删除流程。
- 构建失败不影响当前 published 版本；重复任务不产生重复 Chunk 或索引记录。
- 商品解绑、文档删除和 owner 权限变化不会造成越权检索。
- Worker 重启和多实例并发下任务状态、进度和最终产物一致。

**完成状态**：

- [ ] Phase 04 完成

### Phase 05：基础 RAG 检索与问答能力

**阶段目标**：形成可供 API 和 Agent 复用的基础检索、上下文与带证据问答能力。

**前置依赖**：Phase 04 完成。

**核心开发任务**：

- [ ] 实现 RetrievalScope、Metadata Filtering、Dense Retrieval 和基础 Keyword Retrieval。
- [ ] 实现 Context Builder、Token 预算、来源引用、去重和低证据拒答。
- [ ] 提供独立 RAG 查询接口和本地只读 RAG Tool。
- [ ] 复用现有模型配置完成带证据回答，记录检索与生成配置快照。
- [ ] 建立权限、提示注入、无结果、超时和 Provider 降级测试。

**验收标准**：

- 相同查询通过 API 和 Agent Tool 得到同一权限范围与可追溯证据。
- 平台、账号、商品作用域组合正确，跨 owner 和跨商品测试全部被拦截。
- 证据不足、模型或 Retriever 失败时返回明确降级状态，不编造成功结果。
- 轨迹可定位每次回答使用的索引版本、证据和耗时。

**完成状态**：

- [ ] Phase 05 完成

### Phase 06：Advanced RAG 检索能力

**阶段目标**：在可测量基线上加入高级召回、融合和重排能力。

**前置依赖**：Phase 05 完成，并具备初始标注查询集。

**核心开发任务**：

- [ ] 实现可配置 Query Rewrite 和多查询召回。
- [ ] 完善 Keyword Retrieval、Dense/Keyword 并行执行和 RRF 融合。
- [ ] 实现 Parent-Child Retrieval、结构化 Chunk 扩展和上下文多样性控制。
- [ ] 实现可选 Reranker、超时降级和批量限制。
- [ ] 将全部策略和参数版本化为可复现实验的 Retrieval Profile。

**验收标准**：

- 每项高级能力都有相对基础检索的独立消融结果，不以主观示例作为结论。
- Reranker、Rewrite 超时或不可用时可降级到基础检索，并保留明确轨迹。
- RRF、Parent-Child 和上下文裁剪在固定输入下结果可复现。
- 只有在评估收益和延迟成本可接受时，能力才进入默认 Profile。

**完成状态**：

- [ ] Phase 06 完成

### Phase 07：知识库管理前端与检索实验室

**阶段目标**：在 `agent-web` 提供完整知识管理、任务诊断和检索策略对比界面。

**前置依赖**：Phase 04、Phase 05 完成；Phase 06 的已实现策略可逐步接入。

**核心开发任务**：

- [ ] 增加知识库、文档、版本、作用域、商品绑定和切片预览页面。
- [ ] 增加构建任务进度、失败详情、重试、重建、发布和回滚操作。
- [ ] 增加检索实验室，支持查询、Scope、Profile、TopK、过滤和多方案并排对比。
- [ ] 展示候选来源、各阶段分数、最终证据、耗时、Token 和配置快照。
- [ ] 补齐权限、加载、空、错误、响应式和危险操作确认，并完成真实浏览器验证。

**验收标准**：

- 管理员和普通用户看到并能操作的知识范围符合后端权限。
- 长任务刷新页面后可恢复进度；失败原因和下一步操作清晰可见。
- 实验对比使用真实后端数据，不在正式页面使用 Mock。
- `npm run check` 通过，上传、绑定、构建、重试、实验对比等重要流程通过浏览器检查。

**完成状态**：

- [ ] Phase 07 完成

### Phase 08：智能客服 Agent 与闲鱼业务集成

**阶段目标**：把 RAG 与现有闲鱼消息、人工接管、Tool 和发送任务连接为安全客服闭环。

**前置依赖**：Phase 05 完成；建议 Phase 06、Phase 07 的核心能力已稳定。

**核心开发任务**：

- [ ] 定义客服 Agent Profile、Prompt、上下文、工具白名单、预算和降级策略，与运营 Agent 隔离。
- [ ] 在现有消息入库事务中创建幂等客服任务和 Outbox，实现 Worker 端账号 + 会话级串行、取消和恢复。
- [ ] 实现意图识别、基础 RAG Workflow、Agentic RAG 和静态知识/实时 Tool 路由。
- [ ] 实现回复 ready 事件，复用 Server 端人工接管、回复任务和 WebSocket 发送，增加发送前复核、低可信转人工和重复回复防护。
- [ ] 扩展 Agent 轨迹，记录检索证据、模型、Tool、审批、Token、耗时和终止原因。

**验收标准**：

- 同一触发消息只产生一个客服任务和最多一次已确认回复；同会话消息按顺序处理。
- 人工接管后未发送任务停止，发送前接管竞态测试不会漏发。
- 静态问题只引用有权限的知识；实时问题不使用陈旧向量替代 Tool。
- 敏感写操作未经审批不执行；预算耗尽、低可信和依赖失败均安全降级。
- 真实账号和真实交易验收仍需授权环境，模拟测试不得标记为真实业务验收。

**完成状态**：

- [ ] Phase 08 完成

### Phase 09：RAG Evaluation 与 Benchmark

**阶段目标**：建立可复现、可比较的检索和生成质量评估体系。

**前置依赖**：Phase 05 完成；Advanced RAG 和客服 Agent 可按完成情况接入。

**核心开发任务**：

- [ ] 建立版本化评估数据集、标注流程、导入导出和质量抽检。
- [ ] 实现 Recall@K、Precision@K、MRR、NDCG、Faithfulness 和工程指标计算。
- [ ] 实现实验运行器，比较 Embedding、Chunking、Retriever、Reranker、VectorStore 和参数。
- [ ] 支持消融实验、固定 Benchmark、结果对比、复跑和报告导出。
- [ ] 建立 LLM Judge 版本记录与人工校准流程。

**验收标准**：

- 同一数据、配置和代码版本的实验可复跑并得到可解释的一致结果。
- 每个指标有测试样例验证计算正确性；Faithfulness 结果包含 Judge 与 Prompt 版本。
- 对比报告同时展示质量、延迟、吞吐、Token 和成本，不只展示单一指标。
- 默认检索 Profile 和客服自动发送阈值的选择都有 Benchmark 依据。

**完成状态**：

- [ ] Phase 09 完成

### Phase 10：可观测性、安全性及可靠性完善

**阶段目标**：补齐生产运行所需的指标、告警、安全审计、容量和故障恢复能力。

**前置依赖**：Phase 02 至 Phase 09 的目标能力已基本完成。

**核心开发任务**：

- [ ] 完善构建、检索、Agent、RabbitMQ、Provider 和资源消耗指标及 Dashboard。
- [ ] 增加 OpenTelemetry Trace 与现有 Agent/Event、任务和消息关联标识。
- [ ] 完成凭据、日志、轨迹、Prompt、文档内容和错误消息的脱敏审计。
- [ ] 完成提示注入、越权检索、恶意文档、超大文件、消息风暴和资源耗尽测试。
- [ ] 完成备份恢复、死信处置、索引重建、Provider 故障、限流和容量压测。

**验收标准**：

- 一次客服回答可从消息追踪到任务、检索、模型、Tool、审批和最终发送，且不泄露敏感信息。
- owner、账号、商品隔离的自动化安全测试通过。
- Worker、RabbitMQ、Embedding、VectorStore 和 RustFS 故障演练结果符合已记录的恢复策略。
- 基于真实压测数据形成容量基线、SLO 和告警阈值，不使用估算值冒充结果。

**完成状态**：

- [ ] Phase 10 完成

### Phase 11：集成测试、部署验证及开源文档完善

**阶段目标**：完成端到端验证、可重复部署、升级回退和开源交付说明。

**前置依赖**：Phase 01 至 Phase 10 完成。

**核心开发任务**：

- [ ] 建立从上传、构建、发布、检索、客服回复到轨迹的端到端测试。
- [ ] 验证空库初始化、现有库 Goose 升级、备份恢复、回滚和对象/索引清理。
- [ ] 验证默认 pgvector Compose 与可选 Milvus Profile 的构建、健康和服务可用性。
- [ ] 完成配置、部署、运维、故障排查、数据迁移、评估和安全边界文档。
- [ ] 更新架构、项目范围、数据库、产品、第三方许可和 README 当前事实。

**验收标准**：

- 全部相关单元、集成、前端和端到端检查通过，未通过项有明确阻塞说明。
- 新环境可按文档启动默认栈并完成知识构建和检索；Milvus Profile 可独立验证。
- 已有数据库可无损升级，失败可按文档恢复；不重放 `init.sql`，不执行生产 AutoMigrate。
- 开源仓库不包含真实账号、凭据、会话、业务数据或不可公开测试样本。

**完成状态**：

- [ ] Phase 11 完成

## 14. 开发进度跟踪

| 阶段 | 状态 | 完成时间 | 验收记录 | 遗留问题 |
| --- | --- | --- | --- | --- |
| Phase 01 基础架构与领域模型设计 | 进行中 | - | pgvector、Milvus、Embedding、评估指标及 Tika 中文 OCR 金标准已验证，Parser 首个实现已锁定 | 公开集许可/基线、业务 Qrels 和自建金标准人工核验仍未完成 |
| Phase 02 多进程、消息队列与异步任务 | 未开始 | - | - | - |
| Phase 03 VectorStore 与双实现 | 未开始 | - | - | - |
| Phase 04 知识管理与构建 Pipeline | 未开始 | - | - | - |
| Phase 05 基础 RAG | 未开始 | - | - | - |
| Phase 06 Advanced RAG | 未开始 | - | - | - |
| Phase 07 管理前端与检索实验室 | 未开始 | - | - | - |
| Phase 08 智能客服 Agent | 未开始 | - | - | - |
| Phase 09 Evaluation 与 Benchmark | 未开始 | - | - | - |
| Phase 10 可观测性、安全与可靠性 | 未开始 | - | - | - |
| Phase 11 集成、部署与文档 | 未开始 | - | - | - |

### 当前遗留问题

- EcomRetrieval 数据卡未声明明确许可证；未完成许可确认前不下载、不分发、不产生公开集指标。若维护者无法确认，需替换为许可清晰的公开中文电商检索集。
- 闲鱼模拟集已有 24 篇知识、32 个查询及 45 条 Qrel，但仍待领域人员逐条核验，不得把当前模型对比结果作为生产选型结论。
- 自建复杂文档金标准已完成自动验证，但预期文本、阅读顺序和表格规则仍需领域人员确认；Phase 04 还需补充脱敏真实文档验证泛化。
- 真实闲鱼自动回复验收需要授权账号与明确的业务安全策略。

### 下一步任务

1. 领域人员使用导出的 CSV 核验闲鱼知识、问题和 Qrels，并确认自建复杂文档的文本、表格与阅读顺序预期。
2. 向 EcomRetrieval 维护者确认来源和许可范围；确认后运行公开集基线，无法确认则记录替代数据集决策并运行同口径基线。
3. 汇总人工核验和公开集证据，完成 Phase 01 设计评审；通过后开始 RabbitMQ、Worker 和 Outbox 实现。Reranker 保持 Phase 06 决策门。

## 15. 决策与调整记录

| 日期 | 阶段 | 调整内容 | 原因 | 影响 |
| --- | --- | --- | --- | --- |
| 2026-11-27 | 规划 | 建立初始 11 阶段规划 | 形成后续开发唯一进度基准 | 尚未进入功能实现 |
| 2026-11-27 | Phase 01 | 复用 Eino Parser、Transformer、Embedder，只新增 VectorStore 与高级检索契约 | 现有 Eino 接口已满足单一组件边界，避免重复抽象 | pgvector、Milvus 使用统一项目契约，解析和 Embedding 仍兼容 Eino |
| 2026-11-27 | Phase 01 | PostgreSQL 19 默认向量镜像候选锁定为 pgvector 官方 PG19 镜像 | 临时容器已通过基础兼容性实测 | 正式 Compose 和迁移仍在 Phase 03/11 验证后落地 |
| 2026-11-27 | Phase 01 | Milvus 基础验证采用 3.0.5 Standalone 与官方 Go SDK v3 | 隔离 Compose 实测健康、持久化和权限过滤均通过 | 正式 Compose Profile、备份恢复和容量参数仍在 Phase 03/11 验证后落地 |
| 2026-11-27 | Phase 01 | Apache Tika 3.3.0 作为 PDF/Office Sidecar 候选继续评估 | 合成中文 DOCX/PDF 基础提取通过，但尚无代表性复杂文档支持最终质量结论 | 不提前加入正式 Compose 或生产依赖，待真实样本与 Go 原生方案对比后决定 |
| 2026-11-27 | Phase 01 | 平台知识使用零 owner，账号/商品知识强制使用真实 owner；Provider 错误统一支持 `errors.Is` 分类 | 避免共享知识伪装成用户数据，并让任务层无需匹配错误文本即可决定重建或重试 | VectorStore 实现和契约测试必须覆盖平台共享、租户隔离、索引冲突与暂时不可用 |
| 2026-11-28 | Phase 01 | 本地验证默认使用 Ollama OpenAI Compatible 端点、`qwen3-embedding:0.6b`，以 `bge-m3` 对照；不同模型必须使用独立索引 | 无需外部凭据即可重复验证，两个模型实际输出均为 1024 维 | 只锁定验证基线，不提前锁定生产默认模型；正式选择依赖公开集和人工核验业务集 |
| 2026-11-28 | Phase 01/09 | 统一评估结构并分离通用检索集与闲鱼业务集结论 | 防止公开 Benchmark 与模拟业务数据混淆，所有指标保留数据版本和人工核验状态 | 大型公开集通过脚本下载到 `.codex-work/`；超过 5,000 篇的语料必须走 VectorStore，不使用本地穷举 |
| 2026-11-28 | Phase 01 | Reranker 评估保留在 Phase 06，不作为 Phase 01 阻塞 | Phase 01 只需锁定可插拔边界；Reranker 收益必须建立在稳定召回基线上 | Phase 01 可在 Embedding 与 Parser 验收后结束，不等待 Reranker 凭据 |
| 2026-11-28 | Phase 01 | PDF/Office 首个实现锁定为 Tika 3.3.0 Full Sidecar，并加入固定版本简体中文 OCR 模型 | 单一实现覆盖 Office、文本 PDF 和扫描 PDF；5 类自建金标准按文本、顺序、表格和 OCR 规则全部通过 | Phase 04 实现 Parser Provider；真实文档泛化、资源限制和降级策略继续验收，不把小样本结果表述为生产准确率 |

后续调整必须追加记录，不覆盖历史原因。阶段顺序可以因验证结果调整，但不得跳过实际前置依赖或提前标记完成。
