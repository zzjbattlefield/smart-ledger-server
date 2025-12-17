package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"smart-ledger-server/internal/config"
)

var Log *zap.Logger

// Init 初始化日志
func Init(cfg *config.LogConfig) (*zap.Logger, error) {
	// 设置日志级别
	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// 设置编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 设置编码器
	var encoder zapcore.Encoder
	if cfg.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 设置输出
	writeSyncer, err := getWriteSyncer(cfg)
	if err != nil {
		return nil, err
	}

	// 创建核心
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建Logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	Log = logger

	return Log, nil
}

// getWriteSyncer 获取日志输出
func getWriteSyncer(cfg *config.LogConfig) (zapcore.WriteSyncer, error) {
	// 输出到标准输出
	if cfg.OutputPath == "stdout" || cfg.OutputPath == "" {
		return zapcore.AddSync(os.Stdout), nil
	}

	// 输出到文件，确保目录存在
	dir := filepath.Dir(cfg.OutputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// 使用 lumberjack 实现日志轮转
	lumberJackLogger := &lumberjack.Logger{
		Filename:   cfg.OutputPath,
		MaxSize:    cfg.Rotation.MaxSize,    // 单个日志文件最大大小(MB)
		MaxBackups: cfg.Rotation.MaxBackups, // 保留的旧日志文件数量
		MaxAge:     cfg.Rotation.MaxAge,     // 保留的日志文件最大天数
		Compress:   cfg.Rotation.Compress,   // 是否压缩旧日志文件
	}

	return zapcore.AddSync(lumberJackLogger), nil
}
