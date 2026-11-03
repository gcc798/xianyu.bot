package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gcc798/xianyu.bot/internal/api"
	"github.com/gcc798/xianyu.bot/internal/api/router"
	"github.com/gcc798/xianyu.bot/internal/app"
	"github.com/gcc798/xianyu.bot/internal/config"
	"github.com/gcc798/xianyu.bot/internal/infra"
	"github.com/gcc798/xianyu.bot/internal/infra/mail"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gcc798/xianyu.bot/internal/service"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

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
	if err := infra.Migrate(ctx, db); err != nil {
		return fmt.Errorf("database migration: %w", err)
	}

	mailSender := mail.NewSender(settings.Email)
	services := service.New(ctx, db, mailSender.Send)
	services.RevokedTokens.Load()
	routes := router.New(api.New(services, jwt))
	return infra.RunServer(ctx, app.Routers(ctx, routes, jwt, services), settings.HTTP.Port)
}
