package infra

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/gcc798/go.xianyu.assistant/server/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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
