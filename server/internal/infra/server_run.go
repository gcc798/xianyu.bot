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

type server interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

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
