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

var Logger *zap.Logger
var once sync.Once
var initErr error

// DateWriter 按日期写入器
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

// LevelWriter 按级别写入不同文件的写入器
type LevelWriter struct {
	baseDir     string
	currentDate string
	writers     map[string]*os.File
	mu          sync.Mutex
}

func NewLevelWriter(baseDir string) *LevelWriter {
	return &LevelWriter{
		baseDir: baseDir,
		writers: make(map[string]*os.File),
	}
}

func (w *LevelWriter) Write(level string, p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	dateStr := now.Format("2006-01-02")

	// 如果日期变化，关闭所有旧文件
	if dateStr != w.currentDate {
		for _, file := range w.writers {
			if file != nil {
				file.Close()
			}
		}
		w.writers = make(map[string]*os.File)
		w.currentDate = dateStr
	}

	// 获取或创建对应级别的文件
	writer, exists := w.writers[level]
	if !exists {
		// 创建日期目录
		dateDir := filepath.Join(w.baseDir, dateStr)
		if err := os.MkdirAll(dateDir, 0755); err != nil {
			return 0, fmt.Errorf("failed to create log directory %s: %v", dateDir, err)
		}

		// 创建级别文件
		fileName := filepath.Join(dateDir, level+".log")
		writer, err = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return 0, fmt.Errorf("failed to open log file %s: %v", fileName, err)
		}
		w.writers[level] = writer
	}

	return writer.Write(p)
}

func (w *LevelWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, file := range w.writers {
		if file != nil {
			file.Close()
		}
	}
	return nil
}

// LevelWriteSyncer 适配 zap 的 WriteSyncer
type LevelWriteSyncer struct {
	writer  *LevelWriter
	level   string
	console bool
	mu      sync.Mutex
}

func NewLevelWriteSyncer(writer *LevelWriter, level string, console bool) *LevelWriteSyncer {
	return &LevelWriteSyncer{
		writer:  writer,
		level:   level,
		console: console,
	}
}

func (s *LevelWriteSyncer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 写入对应级别的文件
	n, err = s.writer.Write(s.level, p)
	if err != nil {
		return n, err
	}

	// 如果开启控制台输出，也输出到控制台
	if s.console {
		os.Stdout.Write(p)
	}

	return n, nil
}

func (s *LevelWriteSyncer) Sync() error {
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

	// 如果启用按级别分文件
	if cfg.LevelSeparation {
		levelWriter := NewLevelWriter(filepath.Dir(cfg.OutputPath))

		// 为每个级别创建 core
		cores := []zapcore.Core{}

		levels := []string{"debug", "info", "warn", "error"}
		for _, level := range levels {
			levelEncoder := createEncoder(cfg.Format)
			levelSyncer := NewLevelWriteSyncer(levelWriter, level, cfg.LogInConsole && level == "info")

			var levelCore zapcore.Core
			switch level {
			case "debug":
				levelCore = zapcore.NewCore(levelEncoder, levelSyncer, zapcore.DebugLevel)
			case "info":
				levelCore = zapcore.NewCore(levelEncoder, levelSyncer, zapcore.InfoLevel)
			case "warn":
				levelCore = zapcore.NewCore(levelEncoder, levelSyncer, zapcore.WarnLevel)
			case "error":
				levelCore = zapcore.NewCore(levelEncoder, levelSyncer, zapcore.ErrorLevel)
			}

			cores = append(cores, levelCore)
		}

		// 创建多 core
		core := zapcore.NewTee(cores...)
		Logger = zap.New(core, zap.AddCaller())
	} else {
		// 原有的逻辑
		writeSyncer := createWriteSyncer(cfg)
		core := zapcore.NewCore(encoder, writeSyncer, zapLevel)
		Logger = zap.New(core, zap.AddCaller())
	}

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

// 核心日志函数
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

// Any 添加任意类型的字段
func Any(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}
