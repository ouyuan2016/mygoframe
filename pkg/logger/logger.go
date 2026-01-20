package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"mygoframe/pkg/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Logger *zap.Logger
	once   sync.Once
)

func InitLogger(cfg config.Zap) error {
	var err error
	once.Do(func() {
		encoder := createEncoder(cfg.Format)
		logLevel := parseLogLevel(cfg.Level)

		levelSeparationActive := cfg.RotateByDate && cfg.LevelSeparation

		var core zapcore.Core
		if levelSeparationActive {
			core = createLevelSeparationCore(cfg, encoder, logLevel)
		} else {
			writeSyncer := createWriteSyncer(cfg)
			core = zapcore.NewCore(encoder, writeSyncer, logLevel)
		}

		Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	})
	return err
}

func createEncoder(format string) zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if format == "json" {
		return zapcore.NewJSONEncoder(encoderConfig)
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func createWriteSyncer(cfg config.Zap) zapcore.WriteSyncer {
	var writers []zapcore.WriteSyncer

	if cfg.OutputPath != "" && cfg.OutputPath != "stdout" {
		logPath := cfg.OutputPath
		if cfg.RotateByDate {
			dir := filepath.Dir(logPath)
			filename := filepath.Base(logPath)
			today := time.Now().Format("2006-01-02")
			logPath = filepath.Join(dir, today, filename)
		}

		lumberJackLogger := &lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		writers = append(writers, zapcore.AddSync(lumberJackLogger))
	}

	if cfg.LogInConsole || cfg.OutputPath == "stdout" {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}

	return zapcore.NewMultiWriteSyncer(writers...)
}

func createLevelSeparationCore(cfg config.Zap, encoder zapcore.Encoder, level zapcore.LevelEnabler) zapcore.Core {
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel && level.Enabled(lvl)
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel && level.Enabled(lvl)
	})

	errorWriter := createLevelWriter(cfg, "error")
	infoWriter := createLevelWriter(cfg, "info")

	return zapcore.NewTee(
		zapcore.NewCore(encoder, errorWriter, highPriority),
		zapcore.NewCore(encoder, infoWriter, lowPriority),
	)
}

func createLevelWriter(cfg config.Zap, level string) zapcore.WriteSyncer {
	dir := filepath.Dir(cfg.OutputPath)
	today := time.Now().Format("2006-01-02")
	logFilePath := filepath.Join(dir, today, fmt.Sprintf("%s.log", level))

	lumberJackLogger := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}

func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

func Err(err error) zap.Field {
	return zap.Error(err)
}

func Any(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}
