package core

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Zap() *zap.Logger {
	level, err := zapcore.ParseLevel(global.GVA_CONFIG.Log.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	if global.GVA_CONFIG.Log.Format != "json" {
		config.Encoding = "console"
	}
	options := []zap.Option{zap.AddStacktrace(zapcore.ErrorLevel)}
	if global.GVA_CONFIG.Log.ShowLine {
		options = append(options, zap.AddCaller())
	}
	return zap.Must(config.Build(options...))
}
