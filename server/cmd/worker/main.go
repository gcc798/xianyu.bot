package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gcc798/xianyu.bot/internal/config"
	"github.com/gcc798/xianyu.bot/internal/infra"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

// main 启动独立 RAG 后台进程。
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run 组装 Worker 共享设施；知识构建 Consumer 在对应 Pipeline 落地后注册到此进程。
func run() (runErr error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	settings, err := config.Load()
	if err != nil {
		return fmt.Errorf("加载 Worker 配置: %w", err)
	}
	zap.ReplaceGlobals(infra.Zap(settings.Log))
	db := infra.Postgres(settings.Postgres)
	if db == nil {
		return errors.New("Worker 必须配置 PostgreSQL 数据库")
	}
	if err := infra.MigratePostgres(ctx, db); err != nil {
		return fmt.Errorf("迁移 PostgreSQL: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("读取 PostgreSQL 连接池: %w", err)
	}
	defer func() {
		runErr = errors.Join(runErr, sqlDB.Close())
	}()

	broker, err := infra.OpenRabbitMQ(settings.RabbitMQ)
	if err != nil {
		return err
	}
	defer func() {
		runErr = errors.Join(runErr, broker.Close())
	}()
	if err := broker.DeclareRAGTopology(settings.RabbitMQ); err != nil {
		return err
	}

	readiness := func(checkCtx context.Context) error {
		if err := sqlDB.PingContext(checkCtx); err != nil {
			return fmt.Errorf("PostgreSQL 未就绪: %w", err)
		}
		if err := broker.Health(checkCtx); err != nil {
			return err
		}
		return nil
	}
	workerCtx, cancelWorker := context.WithCancelCause(ctx)
	defer cancelWorker(nil)
	go func() {
		select {
		case fatalErr := <-broker.FatalErrors():
			cancelWorker(fmt.Errorf("RabbitMQ 自动恢复失败: %w", fatalErr))
		case <-workerCtx.Done():
		}
	}()
	zap.L().Info("RAG Worker 启动")
	if err := infra.RunHealthServer(workerCtx, settings.Worker.HealthPort, readiness); err != nil {
		return err
	}
	if cause := context.Cause(workerCtx); cause != nil && !errors.Is(cause, context.Canceled) {
		return cause
	}
	return nil
}
