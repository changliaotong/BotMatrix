package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

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
		zapConfig.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("15:04:05.000"))
		}
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapConfig.EncoderConfig.CallerKey = zapcore.OmitKey
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
	Logger, err = zapConfig.Build(zap.AddCallerSkip(1))
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
	if Logger.Core().Enabled(zap.DebugLevel) {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			callerMsg := fmt.Sprintf("%s [%s:%d]", msg, filepath.Base(file), line)
			Logger.Debug(callerMsg, fields...)
			return
		}
	}
	Logger.Debug(msg, fields...)
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	if Logger.Core().Enabled(zap.InfoLevel) {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			callerMsg := fmt.Sprintf("%s [%s:%d]", msg, filepath.Base(file), line)
			Logger.Info(callerMsg, fields...)
			return
		}
	}
	Logger.Info(msg, fields...)
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	if Logger.Core().Enabled(zap.WarnLevel) {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			callerMsg := fmt.Sprintf("%s [%s:%d]", msg, filepath.Base(file), line)
			Logger.Warn(callerMsg, fields...)
			return
		}
	}
	Logger.Warn(msg, fields...)
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	if Logger.Core().Enabled(zap.ErrorLevel) {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			callerMsg := fmt.Sprintf("%s [%s:%d]", msg, filepath.Base(file), line)
			Logger.Error(callerMsg, fields...)
			return
		}
	}
	Logger.Error(msg, fields...)
}

// Fatal 致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	if Logger.Core().Enabled(zap.FatalLevel) {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			callerMsg := fmt.Sprintf("%s [%s:%d]", msg, filepath.Base(file), line)
			Logger.Fatal(callerMsg, fields...)
			return
		}
	}
	Logger.Fatal(msg, fields...)
}

// Panic 恐慌日志
func Panic(msg string, fields ...zap.Field) {
	if Logger.Core().Enabled(zap.PanicLevel) {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			callerMsg := fmt.Sprintf("%s [%s:%d]", msg, filepath.Base(file), line)
			Logger.Panic(callerMsg, fields...)
			return
		}
	}
	Logger.Panic(msg, fields...)
}

// Sync 同步日志缓冲区
func Sync() error {
	if Logger == nil {
		return nil
	}
	return Logger.Sync()
}

// Printf 兼容标准库 log.Printf
func Printf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	if Logger != nil {
		if Logger.Core().Enabled(zap.InfoLevel) {
			_, file, line, ok := runtime.Caller(1)
			if ok {
				msg = fmt.Sprintf("%s [%s:%d]", msg, filepath.Base(file), line)
			}
		}
		Logger.Info(msg)
	} else {
		fmt.Printf("%s\n", msg)
	}
}

// Println 兼容标准库 log.Println
func Println(v ...any) {
	msg := fmt.Sprintln(v...)
	if Logger != nil {
		if Logger.Core().Enabled(zap.InfoLevel) {
			_, file, line, ok := runtime.Caller(1)
			if ok {
				// Sprintln adds a newline, remove it before appending caller info
				msg = fmt.Sprintf("%s [%s:%d]", msg[:len(msg)-1], filepath.Base(file), line)
			}
		}
		Logger.Info(msg)
	} else {
		fmt.Print(msg)
	}
}

// Fatalf 格式化并记录致命错误
func Fatalf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	if Logger != nil {
		Logger.Fatal(msg)
	} else {
		fmt.Fprintf(os.Stderr, "FATAL: %s\n", msg)
		os.Exit(1)
	}
}

// Errorf 格式化并记录错误
func Errorf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	if Logger != nil {
		Logger.Error(msg)
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", msg)
	}
}

// SetOutput 设置日志输出
func SetOutput(w io.Writer) {
	// 重新构建 Logger 以使用新的输出
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(config),
		zapcore.AddSync(w),
		zap.NewAtomicLevelAt(zapcore.InfoLevel),
	)
	Logger = zap.New(core, zap.AddCallerSkip(1))
	zap.ReplaceGlobals(Logger)
}
