# RAG 技术验证报告

## 1. 结论与边界

本报告记录 2026-11-28 在本机完成的 Embedding、检索指标和文档解析验证。原始 JSON、日志和外部样本保存在被 Git 忽略的 `.codex-work/`，仓库只提交可复现脚本、统一评估结构和结果摘要。

- Ollama OpenAI Compatible Embedding 的连通性、1024 维输出、批量、并发、取消和无效模型处理均已通过。
- Recall@K、Precision@K、MRR、NDCG 已用确定性单元测试和 32 条模拟业务查询端到端验证；结果仍标记为 `pending_human_review`，仅作为开发回归基线，最终核验与调优按用户决定延后至 Phase 09。
- Apache Tika 3.3.0 Full 加固定版本中文 OCR 模型通过 5 类自建金标准样本，作为首个 PDF/Office Parser 实现候选锁定；真实文档泛化仍需在 Phase 04 持续回归。
- EcomRetrieval 数据卡未声明明确许可证。未获得上游许可确认前，不下载、不分发、不生成可对外引用的公开集结论。

## 2. 固定环境与参数

| 项目 | 固定值 |
| --- | --- |
| Ollama | 0.34.0，本机 `http://localhost:11434/v1`；容器配置使用 `http://host.docker.internal:11434/v1` |
| API Key | `ollama`，仅为兼容参数，不是生产凭据 |
| 模型 | `qwen3-embedding:0.6b`、`bge-m3` |
| 维度与距离 | 1024、Cosine |
| 批量与并发 | Batch Size 32，并发 2，共 4 个吞吐请求 |
| 闲鱼模拟集 | `xianyu-business-simulation` v1.1：24 篇知识、32 个查询、45 条 Qrel，`pending_human_review` |
| Apache POI | 提交 `f63b77bb7497d950073969f9d418f2270dd87edd`，每种 Office 格式选择 2 个兼容性样本 |
| Tika | `apache/tika:3.3.0.0-full`；自定义镜像加入固定提交的 `chi_sim.traineddata` |
| 文档金标准 | 自建 DOCX、XLSX、PPTX、多栏中文 PDF、扫描中文 PDF，各 1 个 |

不同 Embedding 模型必须写入独立索引版本，禁止在同一索引混写向量。

## 3. Embedding 实测

### 3.1 运行能力

| 模型 | 实际维度 | 正相关 Cosine | 负相关 Cosine | 分离度 | Batch 32 延迟 | 128 条并发吞吐 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `qwen3-embedding:0.6b` | 1024 | 0.6827 | 0.4279 | 0.2548 | 751.7 ms | 44.23 条/秒 |
| `bge-m3` | 1024 | 0.7763 | 0.5993 | 0.1770 | 628.6 ms | 53.55 条/秒 |

两个模型均通过无效模型拒绝和已取消 Context 的错误处理。本次是单机短时探针，没有预热分层、多轮统计、P95/P99 或资源成本控制，吞吐数据只用于确认运行能力。

### 3.2 闲鱼业务模拟集

| 模型 | Recall@1 | Precision@1 | MRR@1 | NDCG@1 | Recall@3 | MRR@3 | Recall@5 | MRR@5 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `qwen3-embedding:0.6b` | 0.7500 | 0.9375 | 0.9375 | 0.9167 | 0.9063 | 0.9688 | 0.9531 | 0.9688 |
| `bge-m3` | 0.7656 | 0.9688 | 0.9688 | 0.9479 | 0.8750 | 0.9844 | 0.9531 | 0.9844 |

逐查询证据已写入 JSON 报告。Top 1 失败集中在已发货/未发货退款区分和质量问题举证材料等难负例，证明小规模满分结果不能代表业务效果。当前 Qrels 由开发阶段构造，未经过领域人员逐条核验；即使 `bge-m3` 当前数字略高，也不能据此锁定生产默认模型。

人工核验通过 `scripts/rag-eval/review_evaluation_dataset.py` 导出 CSV。脚本要求每条原始 Qrel 都有接受或拒绝结论、记录核验人，并允许补充漏标的有效查询/文档对；只有所有查询仍保留正相关文档时才生成 `human_reviewed` 版本。该流程保留到 Phase 09，由用户结合完整功能和真实使用反馈执行，不阻塞当前功能开发。

## 4. 通用检索集

`scripts/rag-eval/import_ecom_retrieval.py` 固定 EcomRetrieval 提交 `94a400f33e9951f2fa474c1527e29f6a99f7cfe4`，可将 Corpus、Queries、Qrels 转换为项目统一 JSON，并生成原始文件 SHA-256 清单。

本次没有生成 EcomRetrieval 指标：

1. 上游数据卡未声明明确许可证，脚本要求人工审查后显式传入 `--accept-license`；托管平台提供下载能力不等于授予使用权。
2. 未获得维护者或原始权利人的书面许可确认前，保持合规阻塞。
3. 该数据集为 10 万级语料，正式实验必须通过 pgvector 或 Milvus 的同口径 VectorStore 流程执行，不使用本地全量余弦穷举。

解除方式：优先取得维护者对数据来源、许可范围和商业/研究用途的明确确认；若无法取得，替换为许可清晰的公开中文电商检索集，并在决策记录中说明数据分布差异。

## 5. 文档解析实测

### 5.1 固定版本兼容性样本

按固定 Apache POI 提交下载 DOC、DOCX、XLS、XLSX、PPT、PPTX 各 2 个样本。12 个样本中 11 个返回 HTTP 200，`chartex.docx` 返回 422 并保留为回归样本。空 PDF、截断 PDF、截断 DOCX 和加密 PDF 均返回 422，没有静默生成空成功结果。

这些样本只能证明格式兼容性，不能证明解析质量；部分 PPT 返回 200 但正文为空或很少，因此不能把 HTTP 200 当作质量通过。

### 5.2 可验证金标准

`generate_document_goldens.py` 按需生成带预期文本、阅读顺序、表格数量和 OCR 标记的 5 类样本。`benchmark_tika_documents.py` 请求 XHTML，分别检查文本片段召回、阅读顺序和表格数量。

| 样本 | 验收结果 | 单次耗时 | 验收点 |
| --- | --- | ---: | --- |
| DOCX 复杂表格 | 通过 | 335.1 ms | 文本、顺序、1 个表格 |
| XLSX 表格与公式 | 通过 | 31.6 ms | 文本、顺序、1 个表格 |
| PPTX 表格与流程 | 通过 | 177.3 ms | 文本、顺序、1 个表格 |
| 中文多栏 PDF | 通过 | 141.4 ms | 文本与阅读顺序 |
| 中文扫描 PDF | 通过 | 342.4 ms | OCR 文本与顺序 |

最终结果为 5/5 通过，容器观测内存约 325.8 MiB。官方 Full 镜像没有默认安装简体中文模型，项目测试镜像从 `tessdata_fast` 固定提交下载 `chi_sim.traineddata`，并通过 `tika-config.xml` 配置 `chi_sim+eng`。没有启用 `sortByPosition`，因为实测会破坏本组双栏样本的期望阅读顺序。

据此锁定 Tika Full Sidecar 作为 Phase 04 的首个 Parser Provider 实现，不把测试容器直接加入当前生产部署。自建样本的预期值仍需领域人员确认，并继续补充脱敏真实文档、旋转页、跨页表格、手写/低清扫描件和超大文件；5 个样本不能外推为生产解析准确率。

OmniDocBench 的数据用途限制与产品场景不兼容，本轮未下载，也不进入产品数据或商业评估集。

## 6. 可复现入口

```bash
# 业务评估集人工核验表
python3 scripts/rag-eval/review_evaluation_dataset.py export \
  --dataset server/internal/rag/testdata/xianyu-business-v1.json \
  --output .codex-work/rag-eval/xianyu-business-v1.1-review.csv

# 自建复杂文档和金标准
uv run scripts/rag-eval/generate_document_goldens.py \
  --output-dir .codex-work/rag-eval/document-goldens

# 带中文 OCR 的固定 Tika 镜像
docker build -t xianyu-rag-eval-tika:3.3.0.0-chi-sim scripts/rag-eval/tika

# Tika 金标准验证（容器启动后）
python3 scripts/rag-eval/benchmark_tika_documents.py \
  --input-dir .codex-work/rag-eval/document-goldens \
  --golden .codex-work/rag-eval/document-goldens/golden.json \
  --output .codex-work/reports/rag-document-goldens.json

# Ollama 业务评估
cd server
go run ./cmd/rag-eval \
  --dataset internal/rag/testdata/xianyu-business-v1.json \
  --output ../.codex-work/reports/rag-embedding-business.json
```

Phase 09 开始公开集评估、业务 Qrels 人工核验或真实复杂文档回归后，必须同步更新本报告和 `docs/rag-development-plan.md`，不能沿用当前开发基线作为最终效果结论。
