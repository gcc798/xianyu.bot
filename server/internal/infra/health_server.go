package infra

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ReadinessCheck 判断 Worker 的数据库和 Broker 等必要依赖是否可用。
type ReadinessCheck func(context.Context) error

// RunHealthServer 启动 Worker 存活与就绪探针，并随 ctx 取消优雅关闭。
func RunHealthServer(ctx context.Context, port int, readiness ReadinessCheck) error {
	if port <= 0 {
		return fmt.Errorf("Worker 健康端口必须大于零: %d", port)
	}
	address := fmt.Sprintf(":%d", port)
	zap.L().Info("RAG Worker 健康服务启动", zap.String("address", address))
	server := &http.Server{
		Addr:              address,
		Handler:           healthHandler(readiness),
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := runServer(ctx, server); err != nil {
		return err
	}
	zap.L().Info("RAG Worker 健康服务已关闭")
	return nil
}

func healthHandler(readiness ReadinessCheck) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /ready", func(writer http.ResponseWriter, request *http.Request) {
		if readiness == nil {
			http.Error(writer, "not ready", http.StatusServiceUnavailable)
			return
		}
		ctx, cancel := context.WithTimeout(request.Context(), 3*time.Second)
		defer cancel()
		if err := readiness(ctx); err != nil {
			http.Error(writer, "not ready", http.StatusServiceUnavailable)
			return
		}
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("ready"))
	})
	return mux
}
