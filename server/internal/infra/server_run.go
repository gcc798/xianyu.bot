package infra

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// server 描述 HTTP 服务启动与优雅关闭所需的最小能力。
type server interface {
	// ListenAndServe 阻塞运行服务直至失败或关闭。
	ListenAndServe() error
	// Shutdown 使用给定上下文停止接收请求并等待连接退出。
	Shutdown(context.Context) error
}

// RunServer 在 port 上启动 Gin 路由，并随 ctx 取消而优雅关闭。
func RunServer(ctx context.Context, router *gin.Engine, port int) error {
	address := fmt.Sprintf(":%d", port)
	zap.L().Info("闲鱼 Harness Agent 启动", zap.String("address", address))
	srv := &http.Server{
		Addr:           address,
		Handler:        router,
		ReadTimeout:    10 * time.Minute,
		WriteTimeout:   10 * time.Minute,
		MaxHeaderBytes: 1 << 20,
	}
	return runServer(ctx, srv)
}

// runServer 协调服务运行错误和调用方取消信号。
func runServer(ctx context.Context, srv server) error {
	serverErrors := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- fmt.Errorf("listen: %w", err)
		}
	}()

	select {
	case err := <-serverErrors:
		return err
	case <-ctx.Done():
	}
	zap.L().Info("关闭WEB服务...")

	shutdownContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownContext); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	zap.L().Info("WEB服务已关闭")
	return nil
}
