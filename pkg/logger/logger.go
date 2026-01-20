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

// FileWriter 通用的文件写入器，支持按日期分割和按级别分离
type FileWriter struct {
	basePath        string
	currentDate     string
	level           string
	file            *os.File
	rotateByDate    bool
	levelSeparation bool
	mu              sync.Mutex
}

func NewFileWriter(basePath string, level string, rotateByDate bool, levelSeparation bool) *FileWriter {
	return &FileWriter{
		basePath:        basePath,
		level:           level,
		rotateByDate:    rotateByDate,
		levelSeparation: levelSeparation,
	}
}

func (w *FileWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	dateStr := now.Format("2006-01-02")

	// 检查是否需要重新打开文件（日期变化或文件未打开）
	if dateStr != w.currentDate || w.file == nil {
		if w.file != nil {
			w.file.Close()
		}

		fileName := w.getFileName(dateStr)
		dir := filepath.Dir(fileName)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return 0, fmt.Errorf("failed to create log directory: %v", err)
		}

		w.file, err = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return 0, fmt.Errorf("failed to open log file %s: %v", fileName, err)
		}

		w.currentDate = dateStr
	}

	return w.file.Write(p)
}

func (w *FileWriter) getFileName(dateStr string) string {
	if !w.rotateByDate && !w.levelSeparation {
		// 既不按日期也不按级别分离，使用原始路径
		return w.basePath
	}

	if w.levelSeparation {
		// 按级别分离日志文件
		if w.rotateByDate {
			// 按日期和级别分离：log/2026-01-20/info.log
			return filepath.Join(filepath.Dir(w.basePath), dateStr, w.level+".log")
		} else {
			// 只按级别分离：log/info.log
			return filepath.Join(filepath.Dir(w.basePath), w.level+".log")
		}
	} else {
		// 只按日期分离：log/starlive-2026-01-20.log
		ext := filepath.Ext(w.basePath)
		base := w.basePath[:len(w.basePath)-len(ext)]
		return fmt.Sprintf("%s-%s%s", base, dateStr, ext)
	}
}

func (w *FileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// WriteSyncer 适配 zap 的 WriteSyncer
type WriteSyncer struct {
	writer  *FileWriter
	console bool
	mu      sync.Mutex
}

func NewWriteSyncer(writer *FileWriter, console bool) *WriteSyncer {
	return &WriteSyncer{
		writer:  writer,
		console: console,
	}
}

func (s *WriteSyncer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 写入文件
	n, err = s.writer.Write(p)
	if err != nil {
		return n, err
	}

	// 如果开启控制台输出，也输出到控制台
	if s.console {
		os.Stdout.Write(p)
	}

	return n, nil
}

func (s *WriteSyncer) Sync() error {
	return nil
}

func InitLogger(cfg config.Zap) error {
	zapLevel := parseLogLevel(cfg.Level)
	encoder := createEncoder(cfg.Format)

	// 注意：当 rotate-by-date 为 false 时，level-separation 无论是什么值都会失效
	// 这意味着如果不需要按日期分割，那么按级别分离也会失效，使用单一的日志文件
	if !cfg.RotateByDate {
		// 不启用日期分割，使用单一文件模式
		writeSyncer := createWriteSyncer(cfg)
		core := zapcore.NewCore(encoder, writeSyncer, zapLevel)
		Logger = zap.New(core, zap.AddCaller())
		return nil
	}

	// 启用了日期分割
	if cfg.LevelSeparation {
		// 同时启用日期分割和级别分离
		return initLevelSeparationLogger(cfg, zapLevel, encoder)
	} else {
		// 只启用日期分割，不启用级别分离
		writeSyncer := createWriteSyncer(cfg)
		core := zapcore.NewCore(encoder, writeSyncer, zapLevel)
		Logger = zap.New(core, zap.AddCaller())
		return nil
	}
}

// initLevelSeparationLogger 初始化按级别分离的日志器
func initLevelSeparationLogger(cfg config.Zap, zapLevel zapcore.Level, encoder zapcore.Encoder) error {
	// 为每个级别创建 core
	var cores []zapcore.Core
	levels := []string{"debug", "info", "warn", "error"}

	for _, level := range levels {
		levelWriter := NewFileWriter(cfg.OutputPath, level, cfg.RotateByDate, cfg.LevelSeparation)
		levelSyncer := NewWriteSyncer(levelWriter, cfg.LogInConsole && level == "info")

		var levelCore zapcore.Core
		switch level {
		case "debug":
			levelCore = zapcore.NewCore(encoder, levelSyncer, zapcore.DebugLevel)
		case "info":
			levelCore = zapcore.NewCore(encoder, levelSyncer, zapcore.InfoLevel)
		case "warn":
			levelCore = zapcore.NewCore(encoder, levelSyncer, zapcore.WarnLevel)
		case "error":
			levelCore = zapcore.NewCore(encoder, levelSyncer, zapcore.ErrorLevel)
		}

		cores = append(cores, levelCore)
	}

	// 创建多 core
	core := zapcore.NewTee(cores...)
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
	if cfg.OutputPath == "stdout" {
		return zapcore.AddSync(os.Stdout)
	}

	// 如果启用了日期分割，使用 FileWriter
	if cfg.RotateByDate && !cfg.LevelSeparation {
		fileWriter := NewFileWriter(cfg.OutputPath, "", cfg.RotateByDate, cfg.LevelSeparation)
		return zapcore.AddSync(NewWriteSyncer(fileWriter, false))
	}

	// 使用 lumberjack 进行日志轮转
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
