package rag

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestXianyuBusinessEvaluationFixture(t *testing.T) {
	file, err := os.Open("testdata/xianyu-business-v1.json")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, file.Close()) })

	dataset, err := DecodeEvaluationDataset(file)
	require.NoError(t, err)
	require.Equal(t, "business_simulation", dataset.Kind)
	require.Equal(t, "pending_human_review", dataset.ReviewStatus)
	require.Len(t, dataset.Corpus, 24)
	require.Len(t, dataset.Queries, 32)
	require.Len(t, dataset.Qrels, 45)
}

func TestEvaluationDatasetValidate(t *testing.T) {
	dataset := EvaluationDataset{
		SchemaVersion: "1", Name: "xianyu", Version: "v1", Kind: "business_simulation",
		Source: "synthetic", License: "project fixture", ReviewStatus: "pending_human_review",
		Corpus:  []EvaluationDocument{{ID: "d1", Text: "七天退货"}},
		Queries: []EvaluationQuery{{ID: "q1", Text: "可以退货吗"}},
		Qrels:   []RelevanceJudgment{{QueryID: "q1", DocumentID: "d1", Relevance: 1}},
	}
	require.NoError(t, dataset.Validate())

	dataset.Qrels[0].DocumentID = "missing"
	require.ErrorContains(t, dataset.Validate(), "不存在的文档")
}

func TestEvaluationDatasetValidateRequiresCompleteUniqueQrels(t *testing.T) {
	dataset := EvaluationDataset{
		SchemaVersion: "1", Name: "xianyu", Version: "v1", Kind: "business_simulation",
		Source: "synthetic", License: "project fixture", ReviewStatus: "pending_human_review",
		Corpus: []EvaluationDocument{{ID: "d1", Text: "七天退货"}},
		Queries: []EvaluationQuery{
			{ID: "q1", Text: "可以退货吗"},
			{ID: "q2", Text: "还没发货能退款吗"},
		},
		Qrels: []RelevanceJudgment{{QueryID: "q1", DocumentID: "d1", Relevance: 1}},
	}
	require.ErrorContains(t, dataset.Validate(), "评估查询缺少 Qrel: q2")

	dataset.Queries = dataset.Queries[:1]
	dataset.Qrels = append(dataset.Qrels, dataset.Qrels[0])
	require.ErrorContains(t, dataset.Validate(), "重复查询文档对")
}

func TestEvaluationDatasetValidateRequiresReviewerAfterHumanReview(t *testing.T) {
	dataset := EvaluationDataset{
		SchemaVersion: "1", Name: "xianyu", Version: "v1", Kind: "business_simulation",
		Source: "synthetic", License: "project fixture", ReviewStatus: "human_reviewed",
		Corpus:  []EvaluationDocument{{ID: "d1", Text: "七天退货"}},
		Queries: []EvaluationQuery{{ID: "q1", Text: "可以退货吗"}},
		Qrels:   []RelevanceJudgment{{QueryID: "q1", DocumentID: "d1", Relevance: 1}},
	}

	require.EqualError(t, dataset.Validate(), "人工核验完成的数据必须记录核验人")

	dataset.Reviewers = []string{"domain-owner"}
	require.NoError(t, dataset.Validate())
}

func TestDecodeEvaluationDatasetRejectsInvalidJSON(t *testing.T) {
	_, err := DecodeEvaluationDataset(strings.NewReader(`{"name":`))
	require.ErrorContains(t, err, "解析 RAG 评估数据")

	_, err = DecodeEvaluationDataset(strings.NewReader(`{} {}`))
	require.ErrorContains(t, err, "多个 JSON 值")
}

func TestEvaluateRetrieval(t *testing.T) {
	qrels := []RelevanceJudgment{
		{QueryID: "q1", DocumentID: "d1", Relevance: 1},
		{QueryID: "q1", DocumentID: "d2", Relevance: 1},
		{QueryID: "q2", DocumentID: "d4", Relevance: 2},
	}
	rankings := map[string][]RankedDocument{
		"q1": {{DocumentID: "d1", Score: 0.9}, {DocumentID: "d3", Score: 0.8}, {DocumentID: "d2", Score: 0.7}},
		"q2": {{DocumentID: "d5", Score: 0.9}, {DocumentID: "d4", Score: 0.8}},
	}

	metrics, err := EvaluateRetrieval(qrels, rankings, 2)
	require.NoError(t, err)
	require.Equal(t, 2, metrics.QueryCount)
	require.InDelta(t, 0.75, metrics.Recall, 1e-9)
	require.InDelta(t, 0.5, metrics.Precision, 1e-9)
	require.InDelta(t, 0.75, metrics.MRR, 1e-9)
	require.InDelta(t, 0.622038473168458, metrics.NDCG, 1e-9)
}

func TestEvaluateRetrievalIgnoresDuplicateCandidates(t *testing.T) {
	metrics, err := EvaluateRetrieval(
		[]RelevanceJudgment{{QueryID: "q1", DocumentID: "d1", Relevance: 1}},
		map[string][]RankedDocument{"q1": {{DocumentID: "d1"}, {DocumentID: "d1"}, {DocumentID: "d2"}}},
		2,
	)
	require.NoError(t, err)
	require.Equal(t, 1.0, metrics.Recall)
	require.Equal(t, 0.5, metrics.Precision)
}
