package core

import (
	"fmt"
	"os"

	"github.com/flipped-aurora/gin-vue-admin/server/core/internal"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Zap 获取 zap.Logger
// Author [SliverHorn](https://github.com/SliverHorn)
func Zap() (logger *zap.Logger) {
	if err := os.MkdirAll(global.GVA_CONFIG.Zap.Director, 0o755); err != nil {
		fmt.Printf("create log directory: %v\n", err)
	}
	levels := global.GVA_CONFIG.Zap.Levels()
	length := len(levels)
	cores := make([]zapcore.Core, 0, length)
	for i := 0; i < length; i++ {
		core := internal.NewZapCore(levels[i])
		cores = append(cores, core)
	}
	// 构建按级别切分的日志输出。
	logger = zap.New(zapcore.NewTee(cores...))
	// 启用 Error 及以上级别的堆栈捕捉，确保 entry.Stack 可用
	opts := []zap.Option{zap.AddStacktrace(zapcore.ErrorLevel)}
	if global.GVA_CONFIG.Zap.ShowLine {
		opts = append(opts, zap.AddCaller())
	}
	logger = logger.WithOptions(opts...)
	return logger
}
