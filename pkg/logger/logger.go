package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"mygoframe/pkg/config"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger
var once sync.Once
var initErr error

type DateWriter struct {
	basePath    string
	currentDate string
	file        *os.File
	mu          sync.Mutex
}

func NewDateWriter(basePath string) *DateWriter {
	return &DateWriter{
		basePath: basePath,
	}
}

func (w *DateWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	dateStr := now.Format("2006-01-02")

	if dateStr != w.currentDate {
		if w.file != nil {
			w.file.Close()
		}

		dir := filepath.Dir(w.basePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return 0, fmt.Errorf("failed to create log directory: %v", err)
		}

		ext := filepath.Ext(w.basePath)
		base := w.basePath[:len(w.basePath)-len(ext)]
		fileName := fmt.Sprintf("%s-%s%s", base, dateStr, ext)

		w.file, err = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return 0, fmt.Errorf("failed to open log file: %v", err)
		}

		w.currentDate = dateStr
	}

	return w.file.Write(p)
}

func (w *DateWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func InitLogger(cfg config.Zap) error {
	once.Do(func() {
		initErr = initLoggerOnce(cfg)
	})
	return initErr
}

func initLoggerOnce(cfg config.Zap) error {
	zapLevel := parseLogLevel(cfg.Level)
	encoder := createEncoder(cfg.Format)
	writeSyncer := createWriteSyncer(cfg)

	core := zapcore.NewCore(encoder, writeSyncer, zapLevel)
	Logger = zap.New(core, zap.AddCaller())

	return nil
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
	fileSyncer := createFileSyncer(cfg)

	if cfg.LogInConsole {
		consoleSyncer := zapcore.AddSync(os.Stdout)
		return zapcore.NewMultiWriteSyncer(fileSyncer, consoleSyncer)
	}

	return fileSyncer
}

func createFileSyncer(cfg config.Zap) zapcore.WriteSyncer {
	if cfg.RotateByDate {
		dateWriter := NewDateWriter(cfg.OutputPath)
		return zapcore.AddSync(dateWriter)
	}

	if cfg.OutputPath == "stdout" {
		return zapcore.AddSync(os.Stdout)
	}

	lumberJackLogger := &lumberjack.Logger{
		Filename:   cfg.OutputPath,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}

	return zapcore.AddSync(lumberJackLogger)
}

func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

func Info(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	}
}

func Error(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, fields...)
	}
}

func Warn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	}
}

func Debug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	}
}

func Err(err error) zap.Field {
	return zap.Error(err)
}

func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		if raw != "" {
			path = path + "?" + raw
		}

		Info("Gin request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
		)
	}
}
