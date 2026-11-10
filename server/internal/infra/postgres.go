package infra

import (
	"context"
	"embed"
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

//go:embed migrations/goose/*.sql
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

// MigratePostgres 应用 Goose PostgreSQL 迁移后再启动业务服务。
func MigratePostgres(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
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
