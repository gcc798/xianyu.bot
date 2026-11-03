package infra

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/gcc798/xianyu.bot/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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
