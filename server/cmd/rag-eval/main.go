// Command rag-eval 对 OpenAI Compatible Embedding 服务执行可复现的基础能力和吞吐验证。
package main

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	openaiembedding "github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/gcc798/xianyu.bot/internal/rag"
)

type options struct {
	baseURL     string
	apiKey      string
	models      string
	dimensions  int
	batchSize   int
	concurrency int
	requests    int
	timeout     time.Duration
	output      string
	dataset     string
}

type embeddingReport struct {
	GeneratedAt          time.Time       `json:"generatedAt"`
	BaseURL              string          `json:"baseUrl"`
	Dimensions           int             `json:"dimensions"`
	BatchSize            int             `json:"batchSize"`
	Concurrency          int             `json:"concurrency"`
	Requests             int             `json:"requests"`
	InvalidModelRejected bool            `json:"invalidModelRejected"`
	CancellationHandled  bool            `json:"cancellationHandled"`
	Models               []modelReport   `json:"models"`
	DatasetEvaluations   []datasetReport `json:"datasetEvaluations,omitempty"`
}

type modelReport struct {
	Model                  string  `json:"model"`
	ActualDimensions       int     `json:"actualDimensions"`
	PositiveCosine         float64 `json:"positiveCosine"`
	NegativeCosine         float64 `json:"negativeCosine"`
	SemanticSeparation     float64 `json:"semanticSeparation"`
	BatchLatencyMillis     float64 `json:"batchLatencyMillis"`
	ConcurrentDurationSecs float64 `json:"concurrentDurationSeconds"`
	TextsProcessed         int     `json:"textsProcessed"`
	TextsPerSecond         float64 `json:"textsPerSecond"`
}

type datasetReport struct {
	Dataset        string                 `json:"dataset"`
	DatasetVersion string                 `json:"datasetVersion"`
	ReviewStatus   string                 `json:"reviewStatus"`
	Model          string                 `json:"model"`
	Metrics        []rag.RetrievalMetrics `json:"metrics"`
}

type requestResult struct {
	vectorCount int
	err         error
}

func main() {
	config := parseFlags()
	if err := run(config); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseFlags() options {
	var config options
	flag.StringVar(&config.baseURL, "base-url", "http://localhost:11434/v1", "OpenAI Compatible API Base URL")
	flag.StringVar(&config.apiKey, "api-key", "ollama", "API Key")
	flag.StringVar(&config.models, "models", "qwen3-embedding:0.6b,bge-m3", "逗号分隔的 Embedding 模型")
	flag.IntVar(&config.dimensions, "dimensions", 1024, "期望向量维度")
	flag.IntVar(&config.batchSize, "batch-size", 32, "单次批量文本数量")
	flag.IntVar(&config.concurrency, "concurrency", 2, "并发请求数")
	flag.IntVar(&config.requests, "requests", 4, "吞吐测试总请求数")
	flag.DurationVar(&config.timeout, "timeout", 10*time.Minute, "完整验证超时")
	flag.StringVar(&config.output, "output", "", "可选 JSON 报告输出路径")
	flag.StringVar(&config.dataset, "dataset", "", "可选统一评估数据 JSON 路径")
	flag.Parse()
	return config
}

// run 依次验证连通性、维度、中文语义、批量、并发吞吐及可识别异常。
func run(config options) error {
	if config.dimensions <= 0 || config.batchSize <= 0 || config.concurrency <= 0 || config.requests <= 0 {
		return errors.New("dimensions、batch-size、concurrency 和 requests 必须大于零")
	}
	models := splitModels(config.models)
	if len(models) == 0 {
		return errors.New("至少需要一个 Embedding 模型")
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.timeout)
	defer cancel()
	report := embeddingReport{
		GeneratedAt: time.Now(), BaseURL: config.baseURL, Dimensions: config.dimensions,
		BatchSize: config.batchSize, Concurrency: config.concurrency, Requests: config.requests,
	}
	var dataset *rag.EvaluationDataset
	if config.dataset != "" {
		loaded, err := loadDataset(config.dataset)
		if err != nil {
			return err
		}
		dataset = &loaded
	}

	for _, model := range models {
		embedder, err := newEmbedder(ctx, config, model)
		if err != nil {
			return fmt.Errorf("创建 %s Embedder: %w", model, err)
		}
		modelResult, err := checkModel(ctx, embedder, config, model)
		if err != nil {
			return fmt.Errorf("验证 %s: %w", model, err)
		}
		report.Models = append(report.Models, modelResult)
		if dataset != nil {
			datasetResult, err := evaluateDataset(ctx, embedder, config.batchSize, model, *dataset)
			if err != nil {
				return fmt.Errorf("评估 %s/%s: %w", dataset.Name, model, err)
			}
			report.DatasetEvaluations = append(report.DatasetEvaluations, datasetResult)
		}
	}

	report.InvalidModelRejected = rejectsInvalidModel(ctx, config)
	report.CancellationHandled = handlesCancellation(config, models[0])
	if !report.InvalidModelRejected || !report.CancellationHandled {
		return errors.New("Embedding Provider 未正确处理无效模型或取消请求")
	}
	return writeReport(report, config.output)
}

func loadDataset(path string) (rag.EvaluationDataset, error) {
	file, err := os.Open(path)
	if err != nil {
		return rag.EvaluationDataset{}, fmt.Errorf("打开评估数据: %w", err)
	}
	defer file.Close()
	dataset, err := rag.DecodeEvaluationDataset(file)
	if err != nil {
		return rag.EvaluationDataset{}, err
	}
	return dataset, nil
}

func evaluateDataset(
	ctx context.Context,
	embedder embedding.Embedder,
	batchSize int,
	model string,
	dataset rag.EvaluationDataset,
) (datasetReport, error) {
	const maxBruteForceCorpus = 5_000
	if len(dataset.Corpus) > maxBruteForceCorpus {
		return datasetReport{}, fmt.Errorf(
			"语料数量 %d 超过本地穷举上限 %d，请通过 VectorStore 执行大规模评估",
			len(dataset.Corpus), maxBruteForceCorpus,
		)
	}

	documents := make([]string, len(dataset.Corpus))
	for index, document := range dataset.Corpus {
		documents[index] = strings.TrimSpace(document.Title + "\n" + document.Text)
	}
	documentVectors, err := embedBatches(ctx, embedder, documents, batchSize)
	if err != nil {
		return datasetReport{}, fmt.Errorf("向量化语料: %w", err)
	}

	queries := make([]string, len(dataset.Queries))
	for index, query := range dataset.Queries {
		queries[index] = query.Text
	}
	queryVectors, err := embedBatches(ctx, embedder, queries, batchSize)
	if err != nil {
		return datasetReport{}, fmt.Errorf("向量化查询: %w", err)
	}

	rankings := make(map[string][]rag.RankedDocument, len(dataset.Queries))
	for queryIndex, query := range dataset.Queries {
		candidates := make([]rag.RankedDocument, 0, len(dataset.Corpus))
		for documentIndex, document := range dataset.Corpus {
			score, err := cosine(queryVectors[queryIndex], documentVectors[documentIndex])
			if err != nil {
				return datasetReport{}, err
			}
			candidates = append(candidates, rag.RankedDocument{DocumentID: document.ID, Score: score})
		}
		slices.SortFunc(candidates, func(left, right rag.RankedDocument) int {
			return cmp.Compare(right.Score, left.Score)
		})
		rankings[query.ID] = candidates
	}

	metrics := make([]rag.RetrievalMetrics, 0, 3)
	for _, k := range []int{1, 3, 5} {
		result, err := rag.EvaluateRetrieval(dataset.Qrels, rankings, k)
		if err != nil {
			return datasetReport{}, err
		}
		metrics = append(metrics, result)
	}
	return datasetReport{
		Dataset: dataset.Name, DatasetVersion: dataset.Version,
		ReviewStatus: dataset.ReviewStatus, Model: model, Metrics: metrics,
	}, nil
}

func embedBatches(ctx context.Context, embedder embedding.Embedder, texts []string, batchSize int) ([][]float64, error) {
	vectors := make([][]float64, 0, len(texts))
	for start := 0; start < len(texts); start += batchSize {
		batch, err := embedder.EmbedStrings(ctx, texts[start:min(start+batchSize, len(texts))])
		if err != nil {
			return nil, err
		}
		vectors = append(vectors, batch...)
	}
	if len(vectors) != len(texts) {
		return nil, fmt.Errorf("向量数量不匹配: got %d want %d", len(vectors), len(texts))
	}
	return vectors, nil
}

func newEmbedder(ctx context.Context, config options, model string) (embedding.Embedder, error) {
	encodingFormat := openaiembedding.EmbeddingEncodingFormatFloat
	return openaiembedding.NewEmbedder(ctx, &openaiembedding.EmbeddingConfig{
		BaseURL: config.baseURL, APIKey: config.apiKey, Model: model,
		Dimensions: new(config.dimensions), EncodingFormat: &encodingFormat, Timeout: config.timeout,
	})
}

func checkModel(ctx context.Context, embedder embedding.Embedder, config options, model string) (modelReport, error) {
	semanticTexts := []string{
		"这件商品支持七天无理由退货吗？",
		"买家签收后七天内可以申请退货，定制商品除外。",
		"商品使用顺丰快递，通常两天送达。",
	}
	vectors, err := embedder.EmbedStrings(ctx, semanticTexts)
	if err != nil {
		return modelReport{}, fmt.Errorf("中文语义向量化: %w", err)
	}
	if err := validateVectors(vectors, len(semanticTexts), config.dimensions); err != nil {
		return modelReport{}, err
	}
	positiveCosine, err := cosine(vectors[0], vectors[1])
	if err != nil {
		return modelReport{}, err
	}
	negativeCosine, err := cosine(vectors[0], vectors[2])
	if err != nil {
		return modelReport{}, err
	}
	if positiveCosine <= negativeCosine {
		return modelReport{}, fmt.Errorf("中文语义相似度未通过: positive=%f negative=%f", positiveCosine, negativeCosine)
	}

	batch := makeBatch(config.batchSize)
	batchStarted := time.Now()
	batchVectors, err := embedder.EmbedStrings(ctx, batch)
	if err != nil {
		return modelReport{}, fmt.Errorf("批量向量化: %w", err)
	}
	batchLatency := time.Since(batchStarted)
	if err := validateVectors(batchVectors, config.batchSize, config.dimensions); err != nil {
		return modelReport{}, err
	}

	duration, textsProcessed, err := benchmarkConcurrent(ctx, embedder, batch, config)
	if err != nil {
		return modelReport{}, err
	}
	return modelReport{
		Model: model, ActualDimensions: len(vectors[0]), PositiveCosine: positiveCosine,
		NegativeCosine: negativeCosine, SemanticSeparation: positiveCosine - negativeCosine,
		BatchLatencyMillis:     float64(batchLatency.Microseconds()) / 1000,
		ConcurrentDurationSecs: duration.Seconds(), TextsProcessed: textsProcessed,
		TextsPerSecond: float64(textsProcessed) / duration.Seconds(),
	}, nil
}

func benchmarkConcurrent(ctx context.Context, embedder embedding.Embedder, batch []string, config options) (time.Duration, int, error) {
	jobs := make(chan struct{}, config.requests)
	results := make(chan requestResult, config.requests)
	for range config.requests {
		jobs <- struct{}{}
	}
	close(jobs)

	started := time.Now()
	var workers sync.WaitGroup
	for range config.concurrency {
		workers.Go(func() {
			for range jobs {
				vectors, err := embedder.EmbedStrings(ctx, batch)
				results <- requestResult{vectorCount: len(vectors), err: err}
			}
		})
	}
	workers.Wait()
	close(results)
	duration := time.Since(started)

	textsProcessed := 0
	for result := range results {
		if result.err != nil {
			return 0, 0, fmt.Errorf("并发向量化: %w", result.err)
		}
		if result.vectorCount != config.batchSize {
			return 0, 0, fmt.Errorf("并发向量数量不匹配: got %d want %d", result.vectorCount, config.batchSize)
		}
		textsProcessed += result.vectorCount
	}
	return duration, textsProcessed, nil
}

func rejectsInvalidModel(ctx context.Context, config options) bool {
	embedder, err := newEmbedder(ctx, config, "xianyu-missing-embedding-model")
	if err != nil {
		return true
	}
	_, err = embedder.EmbedStrings(ctx, []string{"异常处理验证"})
	return err != nil
}

func handlesCancellation(config options, model string) bool {
	ctx, cancel := context.WithCancel(context.Background())
	embedder, err := newEmbedder(ctx, config, model)
	if err != nil {
		cancel()
		return false
	}
	cancel()
	_, err = embedder.EmbedStrings(ctx, []string{"取消验证"})
	return err != nil
}

func makeBatch(size int) []string {
	batch := make([]string, size)
	for index := range size {
		batch[index] = fmt.Sprintf("闲鱼商品知识样本 %d：支持验货后确认收货。", index+1)
	}
	return batch
}

func validateVectors(vectors [][]float64, count, dimensions int) error {
	if len(vectors) != count {
		return fmt.Errorf("向量数量不匹配: got %d want %d", len(vectors), count)
	}
	for index, vector := range vectors {
		if len(vector) != dimensions {
			return fmt.Errorf("第 %d 个向量维度不匹配: got %d want %d", index, len(vector), dimensions)
		}
	}
	return nil
}

func cosine(left, right []float64) (float64, error) {
	if len(left) == 0 || len(left) != len(right) {
		return 0, errors.New("余弦相似度要求两个等长非空向量")
	}
	dotProduct, leftNorm, rightNorm := 0.0, 0.0, 0.0
	for index, leftValue := range left {
		rightValue := right[index]
		dotProduct += leftValue * rightValue
		leftNorm += leftValue * leftValue
		rightNorm += rightValue * rightValue
	}
	if leftNorm == 0 || rightNorm == 0 {
		return 0, errors.New("余弦相似度不能处理零向量")
	}
	return dotProduct / (math.Sqrt(leftNorm) * math.Sqrt(rightNorm)), nil
}

func splitModels(value string) []string {
	models := make([]string, 0)
	for model := range strings.SplitSeq(value, ",") {
		if model = strings.TrimSpace(model); model != "" {
			models = append(models, model)
		}
	}
	return models
}

func writeReport(report embeddingReport, output string) error {
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("编码评估报告: %w", err)
	}
	encoded = append(encoded, '\n')
	if output == "" {
		_, err = os.Stdout.Write(encoded)
		return err
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return fmt.Errorf("创建报告目录: %w", err)
	}
	if err := os.WriteFile(output, encoded, 0o644); err != nil {
		return fmt.Errorf("写入评估报告: %w", err)
	}
	fmt.Println(output)
	return nil
}
