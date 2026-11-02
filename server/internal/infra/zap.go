package infra

import (
	"github.com/gcc798/xianyu.bot/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Zap(settings config.Log) *zap.Logger {
	level, err := zapcore.ParseLevel(settings.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	if settings.Format != "json" {
		config.Encoding = "console"
	}
	options := []zap.Option{zap.AddStacktrace(zapcore.ErrorLevel)}
	if settings.ShowLine {
		options = append(options, zap.AddCaller())
	}
	return zap.Must(config.Build(options...))
}
