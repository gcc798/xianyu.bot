// Package rag 定义与闲鱼业务无关的知识构建和检索契约。
package rag

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/document/parser"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
)

var (
	// ErrInvalidRequest 表示调用参数或授权范围无效，重试相同请求不会成功。
	ErrInvalidRequest = errors.New("RAG 请求无效")
	// ErrIndexNotFound 表示指定索引版本不存在或已被清理。
	ErrIndexNotFound = errors.New("RAG 索引不存在")
	// ErrIndexConflict 表示已有索引与请求的模型、维度或距离配置不兼容。
	ErrIndexConflict = errors.New("RAG 索引配置冲突")
	// ErrProviderUnavailable 表示外部 Provider 暂时不可用，调用方可以按任务策略重试。
	ErrProviderUnavailable = errors.New("RAG Provider 暂时不可用")
)

// ScopeType 标识知识可被平台、账号或商品范围使用。
type ScopeType string

const (
	// ScopePlatform 表示所有获准用户均可检索的平台知识。
	ScopePlatform ScopeType = "platform"
	// ScopeAccount 表示只可用于指定闲鱼账号的知识。
	ScopeAccount ScopeType = "account"
	// ScopeProduct 表示只可用于指定账号商品的知识。
	ScopeProduct ScopeType = "product"
)

// DistanceMetric 标识向量索引和查询使用的距离度量。
type DistanceMetric string

const (
	// DistanceCosine 使用余弦距离。
	DistanceCosine DistanceMetric = "cosine"
	// DistanceInnerProduct 使用内积距离。
	DistanceInnerProduct DistanceMetric = "inner_product"
	// DistanceL2 使用欧氏距离。
	DistanceL2 DistanceMetric = "l2"
)

// DocumentParser 复用 Eino 的流式文档解析契约。
type DocumentParser = parser.Parser

// Chunker 复用 Eino Transformer，使切片保留已有来源元数据。
type Chunker = document.Transformer

// Embedder 复用 Eino 的批量文本向量化契约。
type Embedder = embedding.Embedder

// IndexSpec 描述一个不可混写 Embedding 的物理索引版本。
type IndexSpec struct {
	ID             string         // ID 是跨 Provider 稳定的索引版本标识。
	Provider       string         // Provider 是 pgvector 或 milvus 等已注册实现。
	EmbeddingModel string         // EmbeddingModel 是生成本索引向量的模型版本。
	Dimensions     int            // Dimensions 是每条向量固定维度。
	Distance       DistanceMetric // Distance 是建索引和查询必须一致的距离度量。
}

// Validate 拒绝无法建立兼容物理索引的配置。
func (s IndexSpec) Validate() error {
	if strings.TrimSpace(s.ID) == "" {
		return invalidRequest("索引版本标识不能为空")
	}
	if strings.TrimSpace(s.Provider) == "" {
		return invalidRequest("向量库 Provider 不能为空")
	}
	if strings.TrimSpace(s.EmbeddingModel) == "" {
		return invalidRequest("Embedding 模型不能为空")
	}
	if s.Dimensions <= 0 {
		return invalidRequest("Embedding 维度必须大于零")
	}
	if !slices.Contains([]DistanceMetric{DistanceCosine, DistanceInnerProduct, DistanceL2}, s.Distance) {
		return invalidRequestf("不支持的向量距离: %s", s.Distance)
	}
	return nil
}

// RetrievalScope 是业务 Service 解析并授权后的不可变检索范围。
// VectorStore 只能在这些已发布索引版本内查询，不能根据模型输出扩大范围。
type RetrievalScope struct {
	OwnerID         uint     // OwnerID 是本次检索所属的系统用户。
	AccountID       *uint    // AccountID 限制账号级和商品级知识。
	ProductID       *uint    // ProductID 限制商品级知识。
	IndexVersionIDs []string // IndexVersionIDs 是当前请求获准使用的已发布索引版本。
	IncludePlatform bool     // IncludePlatform 决定是否合并平台级知识。
}

// Validate 确保检索始终具备 owner 和已发布版本边界。
func (s RetrievalScope) Validate() error {
	if s.OwnerID == 0 {
		return invalidRequest("检索范围缺少用户归属")
	}
	if s.ProductID != nil && s.AccountID == nil {
		return invalidRequest("商品级检索必须同时限定闲鱼账号")
	}
	if s.AccountID != nil && *s.AccountID == 0 {
		return invalidRequest("闲鱼账号标识必须大于零")
	}
	if s.ProductID != nil && *s.ProductID == 0 {
		return invalidRequest("商品标识必须大于零")
	}
	if len(s.IndexVersionIDs) == 0 {
		return invalidRequest("检索范围缺少已发布索引版本")
	}
	for _, versionID := range s.IndexVersionIDs {
		if strings.TrimSpace(versionID) == "" {
			return invalidRequest("索引版本标识不能为空")
		}
	}
	return nil
}

// MetadataFilter 表示调用方获准使用的精确 Metadata 条件。
// 范围授权字段不通过本结构传入，避免普通过滤器覆盖 owner 和索引版本边界。
type MetadataFilter struct {
	Key    string   // Key 是 Provider 映射中允许过滤的元数据键。
	Values []string // Values 是任一命中即可的精确值集合。
}

// ChunkMetadata 保存跨 Provider 一致的来源、权限和父子关系。
type ChunkMetadata struct {
	OwnerID           uint      // OwnerID 是账号/商品知识所属用户；平台知识固定为零。
	KnowledgeBaseID   uint      // KnowledgeBaseID 是 Chunk 所属知识库。
	DocumentID        uint      // DocumentID 是稳定文档标识。
	DocumentVersionID uint      // DocumentVersionID 是不可变内容版本。
	IndexVersionID    string    // IndexVersionID 是生成本 Chunk 的索引版本。
	Scope             ScopeType // Scope 是平台、账号或商品作用域。
	AccountID         *uint     // AccountID 是账号级和商品级知识的闲鱼账号。
	ProductIDs        []uint    // ProductIDs 是商品级知识获准绑定的商品集合。
	ParentChunkID     string    // ParentChunkID 关联 Parent-Child Retrieval 的父 Chunk。
	SourceURI         string    // SourceURI 是用于证据回溯的内部来源标识。
	Position          int       // Position 是 Chunk 在文档版本中的零基顺序。
}

// Validate 校验写入向量库前必须成立的权限与版本不变量。
func (m ChunkMetadata) Validate() error {
	if m.KnowledgeBaseID == 0 || m.DocumentID == 0 || m.DocumentVersionID == 0 {
		return invalidRequest("Chunk 缺少知识库或文档版本归属")
	}
	if strings.TrimSpace(m.IndexVersionID) == "" {
		return invalidRequest("Chunk 缺少索引版本")
	}
	if m.Position < 0 {
		return invalidRequest("Chunk 顺序不能小于零")
	}
	if strings.TrimSpace(m.SourceURI) == "" {
		return invalidRequest("Chunk 缺少可追溯来源")
	}
	switch m.Scope {
	case ScopePlatform:
		if m.OwnerID != 0 || m.AccountID != nil || len(m.ProductIDs) != 0 {
			return invalidRequest("平台知识不能绑定用户、账号或商品")
		}
	case ScopeAccount:
		if m.OwnerID == 0 || m.AccountID == nil || *m.AccountID == 0 || len(m.ProductIDs) != 0 {
			return invalidRequest("账号知识必须绑定有效用户和账号")
		}
	case ScopeProduct:
		if m.OwnerID == 0 || m.AccountID == nil || *m.AccountID == 0 || len(m.ProductIDs) == 0 {
			return invalidRequest("商品知识必须绑定有效用户、账号和商品")
		}
		for _, productID := range m.ProductIDs {
			if productID == 0 {
				return invalidRequest("商品知识包含无效商品标识")
			}
		}
	default:
		return invalidRequestf("不支持的知识作用域: %s", m.Scope)
	}
	return nil
}

// VectorRecord 是写入 VectorStore 的完整 Chunk 和向量。
type VectorRecord struct {
	Document *schema.Document // Document 保存 Chunk 标识、文本和 Eino 元数据。
	Metadata ChunkMetadata    // Metadata 保存 Provider 必须执行的权限和版本过滤字段。
	Vector   []float64        // Vector 是由 IndexSpec 指定模型生成的稠密向量。
}

// Validate 校验 VectorRecord 与目标索引兼容。
func (r VectorRecord) Validate(index IndexSpec) error {
	if err := index.Validate(); err != nil {
		return err
	}
	if r.Document == nil || strings.TrimSpace(r.Document.ID) == "" || strings.TrimSpace(r.Document.Content) == "" {
		return invalidRequest("向量记录缺少 Chunk 标识或文本")
	}
	if err := r.Metadata.Validate(); err != nil {
		return err
	}
	if r.Metadata.IndexVersionID != index.ID {
		return invalidRequest("Chunk 索引版本与目标索引不一致")
	}
	if len(r.Vector) != index.Dimensions {
		return invalidRequestf("向量维度不匹配: got %d want %d", len(r.Vector), index.Dimensions)
	}
	return nil
}

// DenseSearch 描述一次经过授权的稠密向量检索。
type DenseSearch struct {
	Index   IndexSpec        // Index 是本次查询的物理索引版本。
	Vector  []float64        // Vector 是与 Index 使用同一模型生成的查询向量。
	TopK    int              // TopK 是返回候选上限。
	Scope   RetrievalScope   // Scope 是业务 Service 解析后的权限范围。
	Filters []MetadataFilter // Filters 是 Profile 允许追加的普通元数据过滤条件。
}

// Validate 校验查询的索引、维度、权限与普通过滤器。
func (q DenseSearch) Validate() error {
	if err := q.Index.Validate(); err != nil {
		return err
	}
	if err := q.Scope.Validate(); err != nil {
		return err
	}
	if !slices.Contains(q.Scope.IndexVersionIDs, q.Index.ID) {
		return invalidRequest("目标索引不在本次授权范围内")
	}
	if len(q.Vector) != q.Index.Dimensions {
		return invalidRequestf("查询向量维度不匹配: got %d want %d", len(q.Vector), q.Index.Dimensions)
	}
	if q.TopK <= 0 {
		return invalidRequest("TopK 必须大于零")
	}
	for _, filter := range q.Filters {
		key := strings.ToLower(strings.TrimSpace(filter.Key))
		if key == "" || len(filter.Values) == 0 {
			return invalidRequest("Metadata Filter 缺少键或值")
		}
		if isReservedMetadataFilterKey(key) {
			return invalidRequestf("Metadata Filter 不能覆盖授权字段: %s", key)
		}
		for _, value := range filter.Values {
			if strings.TrimSpace(value) == "" {
				return invalidRequest("Metadata Filter 不能包含空值")
			}
		}
	}
	return nil
}

func invalidRequest(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidRequest, message)
}

func invalidRequestf(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidRequest, fmt.Sprintf(format, args...))
}

func isReservedMetadataFilterKey(key string) bool {
	switch key {
	case "owner_id", "scope", "account_id", "product_id", "index_version_id":
		return true
	default:
		return false
	}
}

// SearchHit 是 Provider 返回的已排序候选及统一分数。
type SearchHit struct {
	Document *schema.Document // Document 是命中的 Chunk 文本。
	Metadata ChunkMetadata    // Metadata 用于二次权限校验和证据追溯。
	Score    float64          // Score 越大表示相关性越高，由 Provider 适配器归一化。
}

// VectorStore 隔离 pgvector 与 Milvus 的物理索引和检索协议。
// Count 用于发布前完整性校验，DeleteIndex 只删除指定不可变版本。
// 实现必须包装本包定义的索引和 Provider 错误，使任务层无需匹配错误文本即可决定重建或重试。
type VectorStore interface {
	// EnsureIndex 确保目标物理索引存在且配置与 IndexSpec 一致。
	EnsureIndex(ctx context.Context, index IndexSpec) error
	// Upsert 幂等写入属于同一不可变索引版本的 Chunk。
	Upsert(ctx context.Context, index IndexSpec, records []VectorRecord) error
	// Search 在授权范围内返回按相关性降序排列的候选。
	Search(ctx context.Context, query DenseSearch) ([]SearchHit, error)
	// Count 返回指定索引版本的记录数，用于发布前完整性校验。
	Count(ctx context.Context, index IndexSpec) (int64, error)
	// DeleteIndex 删除指定版本的全部物理索引数据。
	DeleteIndex(ctx context.Context, index IndexSpec) error
	// Health 检查 Provider 是否可接受读写请求。
	Health(ctx context.Context) error
}

// KeywordRetriever 从 PostgreSQL 关键词索引返回与 Dense Search 相同的候选结构。
type KeywordRetriever interface {
	// Retrieve 在授权范围内执行关键词召回并返回统一候选。
	Retrieve(ctx context.Context, query string, topK int, scope RetrievalScope, filters []MetadataFilter) ([]SearchHit, error)
}

// QueryRewriter 将原始问题转换为一个或多个保持原意的检索查询。
type QueryRewriter interface {
	// Rewrite 返回保持原意的一个或多个检索查询。
	Rewrite(ctx context.Context, query string) ([]string, error)
}

// Reranker 在召回后按查询重新排序并裁剪候选。
type Reranker interface {
	// Rerank 按查询重新排序候选并返回最多 topK 条结果。
	Rerank(ctx context.Context, query string, candidates []SearchHit, topK int) ([]SearchHit, error)
}

// ContextBudget 限制证据数量和近似 Token，调用方应使用模型对应计数器复核。
type ContextBudget struct {
	MaxChunks int // MaxChunks 是上下文最多包含的证据块数量。
	MaxTokens int // MaxTokens 是上下文允许使用的模型 Token 上限。
}

// Evidence 是 Context Builder 输出的可追溯证据。
type Evidence struct {
	ChunkID string        // ChunkID 是证据对应的稳定 Chunk 标识。
	Content string        // Content 是经过预算裁剪的证据文本。
	Score   float64       // Score 是融合或重排后的最终分数。
	Source  ChunkMetadata // Source 保存文档版本和来源位置。
}

// BuiltContext 是供模型使用的上下文和独立证据列表。
type BuiltContext struct {
	Text      string     // Text 是明确标记为不可信证据的模型上下文。
	Evidence  []Evidence // Evidence 用于回答引用、轨迹和前端展示。
	Truncated bool       // Truncated 表示候选因预算被裁剪。
}

// ContextBuilder 把已排序候选转换为受预算约束的模型上下文。
type ContextBuilder interface {
	// Build 把候选转换为带证据引用且受预算约束的模型上下文。
	Build(ctx context.Context, query string, candidates []SearchHit, budget ContextBudget) (BuiltContext, error)
}
