# RAG 技术验证报告

## 1. 结论与边界

本报告记录 2026-11-28 在本机完成的 Embedding、检索指标和文档解析验证。原始 JSON、日志和外部样本保存在被 Git 忽略的 `.codex-work/`，仓库只提交可复现脚本、统一评估结构和结果摘要。

- Ollama OpenAI Compatible Embedding 的连通性、1024 维输出、批量、并发、取消和无效模型处理均已通过。
- Recall@K、Precision@K、MRR、NDCG 已用确定性单元测试和端到端模拟业务集验证。
- Apache Tika 3.3.0 已实测 12 个固定版本 Office 样本和 5 个异常样本，但尚不能证明表格结构、阅读顺序或扫描件识别质量达到生产要求。
- 尚无证据支持 `qwen3-embedding:0.6b` 或 `bge-m3` 在闲鱼业务上更优。6 条模拟查询过小且 Qrels 未完成人工核验。

## 2. 固定环境与参数

| 项目 | 固定值 |
| --- | --- |
| Ollama | 0.34.0，本机 `http://localhost:11434/v1`；容器配置使用 `http://host.docker.internal:11434/v1` |
| API Key | `ollama`，仅为兼容参数，不是生产凭据 |
| 模型 | `qwen3-embedding:0.6b`、`bge-m3` |
| 维度与距离 | 1024、Cosine |
| 批量与并发 | Batch Size 32，并发 2，共 4 个吞吐请求 |
| 闲鱼模拟集 | `xianyu-business-simulation` v1：8 篇知识、6 个查询，`pending_human_review` |
| Apache POI | 提交 `f63b77bb7497d950073969f9d418f2270dd87edd`，每种 Office 格式选择 2 个代表性样本 |
| Tika | `apache/tika:3.3.0.0`，服务报告 Apache Tika 3.3.0 |

不同 Embedding 模型必须写入独立索引版本，禁止在同一索引混写向量。

## 3. Embedding 实测

| 模型 | 实际维度 | 正相关 Cosine | 负相关 Cosine | 分离度 | Batch 32 延迟 | 128 条并发吞吐 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `qwen3-embedding:0.6b` | 1024 | 0.6827 | 0.4279 | 0.2548 | 751.7 ms | 44.23 条/秒 |
| `bge-m3` | 1024 | 0.7763 | 0.5993 | 0.1770 | 628.6 ms | 53.55 条/秒 |

两个模型都通过无效模型拒绝和已取消 Context 的错误处理。本次只执行单机短时探针，没有预热分层、多轮统计、P95/P99 或资源成本控制，因此吞吐数据只用于确认运行能力。

模拟业务集上两个模型的 Recall@1、MRR 和 NDCG@1 均为 1.0；由于只有 6 个查询、每个查询只有一个相关文档且标注待人工核验，该结果不用于模型选择。

## 4. 通用检索集

`scripts/rag-eval/import_ecom_retrieval.py` 固定 EcomRetrieval 提交 `94a400f33e9951f2fa474c1527e29f6a99f7cfe4`，将 Corpus、Queries、Qrels 转换为项目统一 JSON，并生成原始文件 SHA-256 清单。

本次没有生成 EcomRetrieval 指标，原因如下：

1. 上游数据卡未声明明确许可证，脚本要求人工审查后显式传入 `--accept-license`。
2. 当前环境访问 Hugging Face 下载地址时连接被重置，无法完成可校验下载。
3. 该数据集为 10 万级语料，不能使用本地全量余弦穷举；正式实验必须通过 pgvector 或 Milvus 的同口径 VectorStore 流程执行。

## 5. 文档解析实测

按固定 Apache POI 提交下载了 DOC、DOCX、XLS、XLSX、PPT、PPTX 各 2 个样本。Tika 结果：

- 12 个 Office 样本中 11 个返回 HTTP 200；`chartex.docx` 返回 HTTP 422，应作为后续候选解析器回归样本。
- 单文件耗时约 9.7–125.9 ms；首次较大的 DOC 样本约 108 ms。该数据是单次顺序调用，不代表 P95。
- 空 PDF、截断 PDF、截断 DOCX 和加密 PDF 均返回 HTTP 422，没有静默生成空成功结果。
- 内容为纯文本但扩展名为 PDF 的样本返回 HTTP 200 并提取 16 个字符，说明检测主要依据内容。
- 测试结束时 Tika 容器约占 297 MiB 内存；容器已删除，没有留下后台任务。

部分 PPT 虽返回 200，但提取字符数为 0 或很少。由于 POI 测试文件没有配套文本、表格和阅读顺序金标准，不能把 HTTP 200 视为解析质量通过。OmniDocBench 只允许在显式确认研究用途并固定完整提交 SHA 后下载；本次未下载，也未将其作为产品数据。

## 6. 可复现入口

```bash
# 公开电商检索集：许可审查通过后执行
uv run scripts/rag-eval/import_ecom_retrieval.py \
  --output-dir .codex-work/rag-eval/ecom-retrieval \
  --accept-license

# 固定版本 Office 样本（默认不下载 OmniDocBench）
python3 scripts/rag-eval/prepare_document_samples.py \
  --output-dir .codex-work/rag-eval/document-samples

# 异常文档
uv run scripts/rag-eval/generate_abnormal_documents.py \
  --output-dir .codex-work/rag-eval/abnormal-documents

# Ollama 运行时验证与小型评估集
cd server
go run ./cmd/rag-eval \
  --dataset internal/rag/testdata/xianyu-business-v1.json \
  --output ../.codex-work/reports/rag-embedding.json
```

完成公开集下载、业务 Qrels 人工核验或复杂文档金标准建设后，必须更新本报告和 `docs/rag-development-plan.md`，不能沿用本次受限结论。
