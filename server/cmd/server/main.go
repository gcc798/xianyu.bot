package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	systemRouter "github.com/gcc798/xianyu.bot/internal/api/router/system"
	systemAPI "github.com/gcc798/xianyu.bot/internal/api/system"
	"github.com/gcc798/xianyu.bot/internal/app"
	"github.com/gcc798/xianyu.bot/internal/config"
	"github.com/gcc798/xianyu.bot/internal/infra"
	"github.com/gcc798/xianyu.bot/internal/infra/mail"
	"github.com/gcc798/xianyu.bot/internal/security"
	systemService "github.com/gcc798/xianyu.bot/internal/service/system"
	xianyuService "github.com/gcc798/xianyu.bot/internal/service/xianyu"
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

	systemServices := systemService.New(db)
	systemServices.RevokedTokens.Load()
	systemRoutes := systemRouter.New(systemAPI.New(systemServices, jwt))
	mailSender := mail.NewSender(settings.Email)
	xianyuDeps := xianyuService.Dependencies{DB: db, SendNotificationEmail: mailSender.Send}
	return infra.RunServer(ctx, app.Routers(ctx, systemRoutes, jwt, xianyuDeps), settings.HTTP.Port)
}
