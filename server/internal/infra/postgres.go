package infra

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gcc798/xianyu.bot/internal/config"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//go:embed migrations/init.sql migrations/goose/*.sql
var migrationFiles embed.FS

// Postgres 按 settings 建立连接池；未配置数据库名时返回 nil。
func Postgres(settings config.Postgres) *gorm.DB {
	if settings.Database == "" {
		return nil
	}
	db, err := gorm.Open(postgres.Open(settings.DSN()), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "", log.LstdFlags), logger.Config{
			SlowThreshold:        200 * time.Millisecond,
			LogLevel:             postgresLogLevel(settings.LogLevel),
			ParameterizedQueries: true,
		}),
	})
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxIdleConns(settings.MaxIdleConns)
	sqlDB.SetMaxOpenConns(settings.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(settings.ConnMaxLifetime) * time.Second)
	return db
}

// MigratePostgres 初始化空库并应用 Goose PostgreSQL 迁移。
func MigratePostgres(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if err := initializePostgres(ctx, sqlDB); err != nil {
		return err
	}
	files, err := fs.Sub(migrationFiles, "migrations/goose")
	if err != nil {
		return err
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, sqlDB, files)
	if err != nil {
		return err
	}
	_, err = provider.Up(ctx)
	return err
}

func initializePostgres(ctx context.Context, db *sql.DB) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("connect for database initialization: %w", err)
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin database initialization: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('xianyu.bot:init'))`); err != nil {
		return fmt.Errorf("lock database initialization: %w", err)
	}
	var gooseTable bool
	if err := tx.QueryRowContext(ctx, `SELECT to_regclass('public.goose_db_version') IS NOT NULL`).Scan(&gooseTable); err != nil {
		return fmt.Errorf("check Goose history: %w", err)
	}
	if gooseTable {
		var hasHistory bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM public.goose_db_version)`).Scan(&hasHistory); err != nil {
			return fmt.Errorf("read Goose history: %w", err)
		}
		if !hasHistory {
			return fmt.Errorf("database has an empty Goose history; refusing to guess its schema version")
		}
		return tx.Commit()
	}

	var hasObjects bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'public' AND c.relkind IN ('r', 'p', 'v', 'm', 'S', 'f')
	)`).Scan(&hasObjects); err != nil {
		return fmt.Errorf("check database schema: %w", err)
	}
	if hasObjects {
		return fmt.Errorf("database has schema objects but no Goose history; refusing to initialize or guess its version")
	}

	initSQL, err := migrationFiles.ReadFile("migrations/init.sql")
	if err != nil {
		return fmt.Errorf("read database initialization SQL: %w", err)
	}
	if _, err := tx.ExecContext(ctx, string(initSQL)); err != nil {
		return fmt.Errorf("initialize empty database: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit database initialization: %w", err)
	}
	if _, err := conn.ExecContext(ctx, `RESET ALL`); err != nil {
		return fmt.Errorf("reset database connection settings: %w", err)
	}
	return nil
}

// postgresLogLevel 将配置文本转换为 GORM 日志级别，未知值默认使用 Info。
func postgresLogLevel(value string) logger.LogLevel {
	switch strings.ToLower(value) {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	default:
		return logger.Info
	}
}
