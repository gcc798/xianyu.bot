package rag

import (
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/require"
)

func TestIndexSpecValidate(t *testing.T) {
	valid := IndexSpec{ID: "index-v1", Provider: "pgvector", EmbeddingModel: "embedding-v1", Dimensions: 3, Distance: DistanceCosine}
	require.NoError(t, valid.Validate())

	tests := []struct {
		name  string    // name 是测试场景。
		index IndexSpec // index 是待校验索引配置。
	}{
		{name: "missing id", index: IndexSpec{Provider: "pgvector", EmbeddingModel: "embedding-v1", Dimensions: 3, Distance: DistanceCosine}},
		{name: "missing provider", index: IndexSpec{ID: "index-v1", EmbeddingModel: "embedding-v1", Dimensions: 3, Distance: DistanceCosine}},
		{name: "missing model", index: IndexSpec{ID: "index-v1", Provider: "pgvector", Dimensions: 3, Distance: DistanceCosine}},
		{name: "invalid dimensions", index: IndexSpec{ID: "index-v1", Provider: "pgvector", EmbeddingModel: "embedding-v1", Distance: DistanceCosine}},
		{name: "invalid distance", index: IndexSpec{ID: "index-v1", Provider: "pgvector", EmbeddingModel: "embedding-v1", Dimensions: 3, Distance: "unknown"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.ErrorIs(t, test.index.Validate(), ErrInvalidRequest)
		})
	}
}

func TestRetrievalScopeValidate(t *testing.T) {
	accountID, productID, zeroID := uint(10), uint(20), uint(0)
	require.NoError(t, (RetrievalScope{OwnerID: 1, AccountID: &accountID, ProductID: &productID, IndexVersionIDs: []string{"index-v1"}}).Validate())
	require.Error(t, (RetrievalScope{IndexVersionIDs: []string{"index-v1"}}).Validate())
	require.Error(t, (RetrievalScope{OwnerID: 1, ProductID: &productID, IndexVersionIDs: []string{"index-v1"}}).Validate())
	require.Error(t, (RetrievalScope{OwnerID: 1, AccountID: &zeroID, IndexVersionIDs: []string{"index-v1"}}).Validate())
	require.Error(t, (RetrievalScope{OwnerID: 1}).Validate())
	require.Error(t, (RetrievalScope{OwnerID: 1, IndexVersionIDs: []string{""}}).Validate())
}

func TestChunkMetadataValidateScope(t *testing.T) {
	accountID, firstProductID, secondProductID := uint(10), uint(20), uint(21)
	base := ChunkMetadata{OwnerID: 1, KnowledgeBaseID: 2, DocumentID: 3, DocumentVersionID: 4, IndexVersionID: "index-v1", SourceURI: "rustfs://knowledge/doc-1"}

	platform := base
	platform.Scope, platform.OwnerID = ScopePlatform, 0
	require.NoError(t, platform.Validate())

	account := base
	account.Scope, account.AccountID = ScopeAccount, &accountID
	require.NoError(t, account.Validate())

	product := base
	product.Scope, product.AccountID, product.ProductIDs = ScopeProduct, &accountID, []uint{firstProductID, secondProductID}
	require.NoError(t, product.Validate())

	invalidPlatform := platform
	invalidPlatform.OwnerID = 1
	require.Error(t, invalidPlatform.Validate())

	invalidAccount := account
	invalidAccount.ProductIDs = []uint{firstProductID}
	require.Error(t, invalidAccount.Validate())

	invalidProduct := product
	invalidProduct.ProductIDs = nil
	require.Error(t, invalidProduct.Validate())

	missingSource := platform
	missingSource.SourceURI = ""
	require.Error(t, missingSource.Validate())

	invalidProductID := product
	invalidProductID.ProductIDs = []uint{0}
	require.Error(t, invalidProductID.Validate())
}

func TestVectorRecordValidate(t *testing.T) {
	index := IndexSpec{ID: "index-v1", Provider: "pgvector", EmbeddingModel: "embedding-v1", Dimensions: 3, Distance: DistanceCosine}
	record := VectorRecord{
		Document: &schema.Document{ID: "chunk-1", Content: "退换货规则"},
		Metadata: ChunkMetadata{
			KnowledgeBaseID: 2, DocumentID: 3, DocumentVersionID: 4,
			IndexVersionID: index.ID, Scope: ScopePlatform, SourceURI: "text://faq/1",
		},
		Vector: []float64{0.1, 0.2, 0.3},
	}
	require.NoError(t, record.Validate(index))

	invalidDimension := record
	invalidDimension.Vector = []float64{0.1}
	require.Error(t, invalidDimension.Validate(index))

	invalidVersion := record
	invalidVersion.Metadata.IndexVersionID = "index-v2"
	require.Error(t, invalidVersion.Validate(index))
}

func TestDenseSearchValidate(t *testing.T) {
	index := IndexSpec{ID: "index-v1", Provider: "pgvector", EmbeddingModel: "embedding-v1", Dimensions: 3, Distance: DistanceCosine}
	query := DenseSearch{
		Index: index, Vector: []float64{0.1, 0.2, 0.3}, TopK: 5,
		Scope:   RetrievalScope{OwnerID: 1, IndexVersionIDs: []string{"index-v1"}},
		Filters: []MetadataFilter{{Key: "language", Values: []string{"zh-CN"}}},
	}
	require.NoError(t, query.Validate())

	unauthorized := query
	unauthorized.Scope.IndexVersionIDs = []string{"index-v2"}
	require.Error(t, unauthorized.Validate())

	invalidFilter := query
	invalidFilter.Filters = []MetadataFilter{{Key: "language"}}
	require.Error(t, invalidFilter.Validate())

	authorizationOverride := query
	authorizationOverride.Filters = []MetadataFilter{{Key: "owner_id", Values: []string{"2"}}}
	require.Error(t, authorizationOverride.Validate())
}
