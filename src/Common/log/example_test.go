package log

import (
	"testing"

	"go.uber.org/zap"
)

func TestLogger(t *testing.T) {
	// 初始化默认日志
	InitDefaultLogger()
	defer Sync()

	// 测试不同级别的日志
	Debug("Debug message", zap.String("key", "value"))
	Info("Info message", zap.String("module", "test"), zap.Int("version", 1))
	Warn("Warn message", zap.Bool("warning", true))
	Error("Error message", zap.Error(nil))

	// 测试结构化日志
	Info("User login",
		zap.String("user_id", "12345"),
		zap.String("ip", "192.168.1.1"),
		zap.Bool("success", true),
	)

	// 测试自定义配置
	config := Config{
		Level:       "debug",
		Format:      "json",
		OutputPath:  "stdout",
		ErrorPath:   "stderr",
		Development: false,
	}

	if err := InitLogger(config); err != nil {
		Error("Failed to init logger", zap.Error(err))
	}

	Info("JSON format log", zap.String("format", "json"))
}