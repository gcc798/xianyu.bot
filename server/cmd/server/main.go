package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gcc798/xianyu.bot/internal/agent"
	agenttool "github.com/gcc798/xianyu.bot/internal/agent/tool"
	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gcc798/xianyu.bot/internal/api/router"
	"github.com/gcc798/xianyu.bot/internal/config"
	"github.com/gcc798/xianyu.bot/internal/infra"
	"github.com/gcc798/xianyu.bot/internal/infra/mail"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gcc798/xianyu.bot/internal/service"
	"github.com/gcc798/xianyu.bot/internal/xianyu"
	xianyutool "github.com/gcc798/xianyu.bot/internal/xianyu/agenttool"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

// main 启动后端进程。
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run 装配并运行后端服务。
func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	settings, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	jwt, err := security.NewJWT(settings.JWT)
	if err != nil {
		return fmt.Errorf("invalid JWT config: %w", err)
	}

	zap.ReplaceGlobals(infra.Zap(settings.Log))
	db := infra.Postgres(settings.Postgres)
	if db == nil {
		return fmt.Errorf("PostgreSQL database name is required")
	}
	mailSender := mail.NewSender(settings.Email)
	var sendNotificationEmail func(string, string) error
	if settings.Email.To != "" {
		sendNotificationEmail = func(subject string, body string) error {
			return mailSender.Send(settings.Email.To, subject, body)
		}
	}
	services := service.New(ctx, db, sendNotificationEmail)
	services.RevokedTokens.Load()
	if err := agent.Initialize(ctx, db); err != nil {
		return fmt.Errorf("initialize Agent: %w", err)
	}
	if err := xianyu.Initialize(ctx, db, services.WebSocket); err != nil {
		return fmt.Errorf("initialize Xianyu: %w", err)
	}

	events := agent.NewEventHub(db)
	mcpRegistry := agent.NewMCPRegistry(db)
	resources := agent.NewResourceCatalog(db, events, mcpRegistry)
	sessions := agent.NewSession(db)
	aiConfig := agent.NewAIConfig(db)
	localTools := agenttool.NewRegistry()
	if err := localTools.Register(agenttool.NewCurrentTime()); err != nil {
		return fmt.Errorf("register current time Agent tool: %w", err)
	}
	if err := xianyutool.Register(localTools, services.Account, services.Product, services.Order, services.Message); err != nil {
		return fmt.Errorf("register Xianyu Agent tools: %w", err)
	}
	runner := agent.NewRunner(ctx, db, mcpRegistry, resources, events, localTools)
	routes := router.New(api.New(services, jwt, aiConfig, resources, sessions, runner, events))
	return infra.RunServer(ctx, routes.Engine(jwt), settings.HTTP.Port)
}
