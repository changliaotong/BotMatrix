package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger
)

// Config 日志配置

type Config struct {
	Level       string `json:"level"`
	Format      string `json:"format"` // console or json
	OutputPath  string `json:"output_path"`
	ErrorPath   string `json:"error_path"`
	Development bool   `json:"development"`
}

// InitLogger 初始化日志系统
func InitLogger(config Config) error {
	var zapConfig zap.Config

	if config.Development {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	// 设置日志级别
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		return err
	}
	zapConfig.Level.SetLevel(level)

	// 设置输出格式
	if config.Format == "console" {
		zapConfig.Encoding = "console"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapConfig.Encoding = "json"
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// 设置输出路径
	if config.OutputPath != "" {
		zapConfig.OutputPaths = []string{config.OutputPath}
	} else {
		zapConfig.OutputPaths = []string{"stdout"}
	}

	if config.ErrorPath != "" {
		zapConfig.ErrorOutputPaths = []string{config.ErrorPath}
	} else {
		zapConfig.ErrorOutputPaths = []string{"stderr"}
	}

	// 构建日志实例
	Logger, err = zapConfig.Build()
	if err != nil {
		return err
	}

	// 替换全局日志
	zap.ReplaceGlobals(Logger)

	return nil
}

// InitDefaultLogger 使用默认配置初始化日志
func InitDefaultLogger() {
	config := Config{
		Level:       "info",
		Format:      "console",
		OutputPath:  "stdout",
		ErrorPath:   "stderr",
		Development: true,
	}

	if err := InitLogger(config); err != nil {
		panic(err)
	}
}

// Debug 调试日志
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Fatal 致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

// Panic 恐慌日志
func Panic(msg string, fields ...zap.Field) {
	Logger.Panic(msg, fields...)
}

// Sync 同步日志缓冲区
func Sync() error {
	return Logger.Sync()
}
