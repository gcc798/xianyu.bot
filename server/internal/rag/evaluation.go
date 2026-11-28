package rag

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
	"strings"
)

// EvaluationDataset 是通用检索集或闲鱼业务集使用的统一、可版本化评估结构。
type EvaluationDataset struct {
	SchemaVersion string               `json:"schemaVersion"` // SchemaVersion 标识文件结构版本。
	Name          string               `json:"name"`          // Name 是数据集稳定名称。
	Version       string               `json:"version"`       // Version 固定上游提交或人工标注版本。
	Kind          string               `json:"kind"`          // Kind 区分通用检索集和闲鱼业务集。
	Source        string               `json:"source"`        // Source 记录数据来源，不承载凭据。
	License       string               `json:"license"`       // License 记录使用限制，未知时必须明确标注。
	ReviewStatus  string               `json:"reviewStatus"`  // ReviewStatus 标识人工核验是否完成。
	Corpus        []EvaluationDocument `json:"corpus"`        // Corpus 是参与检索的文档集合。
	Queries       []EvaluationQuery    `json:"queries"`       // Queries 是评估查询集合。
	Qrels         []RelevanceJudgment  `json:"qrels"`         // Qrels 是查询与相关文档的人工或公开标注。
}

// EvaluationDocument 是统一评估集中的可检索文档。
type EvaluationDocument struct {
	ID    string `json:"id"`    // ID 是数据集版本内稳定且唯一的文档标识。
	Title string `json:"title"` // Title 是可选标题。
	Text  string `json:"text"`  // Text 是实际参与向量化和检索的正文。
}

// EvaluationQuery 是统一评估集中的检索问题。
type EvaluationQuery struct {
	ID   string `json:"id"`   // ID 是数据集版本内稳定且唯一的查询标识。
	Text string `json:"text"` // Text 是实际参与向量化的查询文本。
}

// RelevanceJudgment 描述查询与文档的分级相关性，零分记录不参与指标计算。
type RelevanceJudgment struct {
	QueryID    string `json:"queryId"`    // QueryID 关联评估查询。
	DocumentID string `json:"documentId"` // DocumentID 关联语料文档。
	Relevance  int    `json:"relevance"`  // Relevance 大于零表示相关，数值越大相关性越高。
}

// RankedDocument 是检索系统为一个查询返回的有序候选。
type RankedDocument struct {
	DocumentID string  `json:"documentId"` // DocumentID 是候选文档标识。
	Score      float64 `json:"score"`      // Score 是检索系统输出的排序分数。
}

// RetrievalMetrics 保存固定 K 下按查询宏平均的真实检索指标。
type RetrievalMetrics struct {
	K          int     `json:"k"`          // K 是参与计算的候选截断位置。
	QueryCount int     `json:"queryCount"` // QueryCount 是拥有正相关标注的查询数量。
	Recall     float64 `json:"recall"`     // Recall 是 Recall@K 的宏平均值。
	Precision  float64 `json:"precision"`  // Precision 是 Precision@K 的宏平均值。
	MRR        float64 `json:"mrr"`        // MRR 是首个相关结果倒数排名的宏平均值。
	NDCG       float64 `json:"ndcg"`       // NDCG 是使用分级相关性计算的 NDCG@K 宏平均值。
}

// DecodeEvaluationDataset 读取并校验统一评估数据，拒绝悬空 Qrel 和重复标识。
func DecodeEvaluationDataset(reader io.Reader) (EvaluationDataset, error) {
	var dataset EvaluationDataset
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&dataset); err != nil {
		return EvaluationDataset{}, fmt.Errorf("解析 RAG 评估数据: %w", err)
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return EvaluationDataset{}, errors.New("RAG 评估数据包含多个 JSON 值")
		}
		return EvaluationDataset{}, fmt.Errorf("解析 RAG 评估数据尾部: %w", err)
	}
	if err := dataset.Validate(); err != nil {
		return EvaluationDataset{}, err
	}
	return dataset, nil
}

// Validate 校验数据集身份、语料、查询和相关性标注的一致性。
func (d EvaluationDataset) Validate() error {
	if strings.TrimSpace(d.SchemaVersion) == "" || strings.TrimSpace(d.Name) == "" || strings.TrimSpace(d.Version) == "" {
		return errors.New("评估数据缺少结构版本、名称或数据版本")
	}
	if strings.TrimSpace(d.Kind) == "" || strings.TrimSpace(d.Source) == "" ||
		strings.TrimSpace(d.License) == "" || strings.TrimSpace(d.ReviewStatus) == "" {
		return errors.New("评估数据缺少类型、来源、许可或核验状态")
	}
	if len(d.Corpus) == 0 || len(d.Queries) == 0 || len(d.Qrels) == 0 {
		return errors.New("评估数据必须同时包含 Corpus、Queries 和 Qrels")
	}

	documentIDs := make(map[string]struct{}, len(d.Corpus))
	for _, document := range d.Corpus {
		if strings.TrimSpace(document.ID) == "" || strings.TrimSpace(document.Text) == "" {
			return errors.New("评估语料包含空标识或空正文")
		}
		if _, exists := documentIDs[document.ID]; exists {
			return fmt.Errorf("评估语料包含重复文档: %s", document.ID)
		}
		documentIDs[document.ID] = struct{}{}
	}

	queryIDs := make(map[string]struct{}, len(d.Queries))
	for _, query := range d.Queries {
		if strings.TrimSpace(query.ID) == "" || strings.TrimSpace(query.Text) == "" {
			return errors.New("评估查询包含空标识或空文本")
		}
		if _, exists := queryIDs[query.ID]; exists {
			return fmt.Errorf("评估数据包含重复查询: %s", query.ID)
		}
		queryIDs[query.ID] = struct{}{}
	}

	for _, qrel := range d.Qrels {
		if qrel.Relevance <= 0 {
			return errors.New("Qrel 相关性必须大于零")
		}
		if _, exists := queryIDs[qrel.QueryID]; !exists {
			return fmt.Errorf("Qrel 引用了不存在的查询: %s", qrel.QueryID)
		}
		if _, exists := documentIDs[qrel.DocumentID]; !exists {
			return fmt.Errorf("Qrel 引用了不存在的文档: %s", qrel.DocumentID)
		}
	}
	return nil
}

// EvaluateRetrieval 按 Qrel 查询宏平均计算 Recall@K、Precision@K、MRR 和 NDCG@K。
// 重复候选只按首次出现计分，缺失候选按不相关结果处理。
func EvaluateRetrieval(qrels []RelevanceJudgment, rankings map[string][]RankedDocument, k int) (RetrievalMetrics, error) {
	if k <= 0 {
		return RetrievalMetrics{}, errors.New("评估 K 必须大于零")
	}
	relevanceByQuery := make(map[string]map[string]int)
	for _, qrel := range qrels {
		if qrel.Relevance <= 0 || strings.TrimSpace(qrel.QueryID) == "" || strings.TrimSpace(qrel.DocumentID) == "" {
			return RetrievalMetrics{}, errors.New("Qrel 必须包含查询、文档和正相关分值")
		}
		if relevanceByQuery[qrel.QueryID] == nil {
			relevanceByQuery[qrel.QueryID] = make(map[string]int)
		}
		relevanceByQuery[qrel.QueryID][qrel.DocumentID] = max(relevanceByQuery[qrel.QueryID][qrel.DocumentID], qrel.Relevance)
	}
	if len(relevanceByQuery) == 0 {
		return RetrievalMetrics{}, errors.New("没有可评估的正相关标注")
	}

	metrics := RetrievalMetrics{K: k, QueryCount: len(relevanceByQuery)}
	for queryID, relevant := range relevanceByQuery {
		ranking := rankings[queryID]
		seen := make(map[string]struct{}, min(k, len(ranking)))
		hits := 0
		reciprocalRank := 0.0
		dcg := 0.0
		rank := 0
		for _, candidate := range ranking {
			if rank == k {
				break
			}
			if _, duplicate := seen[candidate.DocumentID]; duplicate {
				continue
			}
			seen[candidate.DocumentID] = struct{}{}
			rank++
			relevance := relevant[candidate.DocumentID]
			if relevance <= 0 {
				continue
			}
			hits++
			if reciprocalRank == 0 {
				reciprocalRank = 1 / float64(rank)
			}
			dcg += discountedGain(relevance, rank)
		}

		idealRelevance := make([]int, 0, len(relevant))
		for _, relevance := range relevant {
			idealRelevance = append(idealRelevance, relevance)
		}
		slices.Sort(idealRelevance)
		slices.Reverse(idealRelevance)
		idcg := 0.0
		for index, relevance := range idealRelevance[:min(k, len(idealRelevance))] {
			idcg += discountedGain(relevance, index+1)
		}

		metrics.Recall += float64(hits) / float64(len(relevant))
		metrics.Precision += float64(hits) / float64(k)
		metrics.MRR += reciprocalRank
		if idcg > 0 {
			metrics.NDCG += dcg / idcg
		}
	}

	queryCount := float64(metrics.QueryCount)
	metrics.Recall /= queryCount
	metrics.Precision /= queryCount
	metrics.MRR /= queryCount
	metrics.NDCG /= queryCount
	return metrics, nil
}

func discountedGain(relevance, rank int) float64 {
	return (math.Pow(2, float64(relevance)) - 1) / math.Log2(float64(rank)+1)
}
