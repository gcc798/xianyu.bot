package main

import (
	"context"
	"fmt"
	"os"

	systemRouter "github.com/gcc798/go.xianyu.assistant/server/internal/api/router/system"
	systemAPI "github.com/gcc798/go.xianyu.assistant/server/internal/api/system"
	"github.com/gcc798/go.xianyu.assistant/server/internal/app"
	"github.com/gcc798/go.xianyu.assistant/server/internal/config"
	"github.com/gcc798/go.xianyu.assistant/server/internal/database"
	"github.com/gcc798/go.xianyu.assistant/server/internal/infra"
	"github.com/gcc798/go.xianyu.assistant/server/internal/infra/mail"
	"github.com/gcc798/go.xianyu.assistant/server/internal/security"
	systemService "github.com/gcc798/go.xianyu.assistant/server/internal/service/system"
	xianyuService "github.com/gcc798/go.xianyu.assistant/server/internal/service/xianyu"
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
	settings, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	jwt, err := security.NewJWT(settings.JWT)
	if err != nil {
		return fmt.Errorf("invalid JWT config: %w", err)
	}

	infra.Log = infra.Zap(settings.Log)
	zap.ReplaceGlobals(infra.Log)
	infra.DB = infra.Postgres(settings.Postgres)
	if infra.DB == nil {
		return fmt.Errorf("PostgreSQL database name is required")
	}
	if err := database.Migrate(context.Background(), infra.DB); err != nil {
		return fmt.Errorf("database migration: %w", err)
	}

	systemServices := systemService.New(infra.DB)
	systemServices.JWT.LoadRevokedTokens()
	systemRoutes := systemRouter.New(systemAPI.New(systemServices, jwt))
	mailSender := mail.NewSender(settings.Email)
	xianyuDeps := xianyuService.Dependencies{DB: infra.DB, SendNotificationEmail: mailSender.Send}
	infra.RunServer(app.Routers(systemRoutes, jwt, xianyuDeps), settings.HTTP.Port)
	return nil
}
