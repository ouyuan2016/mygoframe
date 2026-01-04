package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger
var once sync.Once
var initErr error

// 不同级别的日志记录器
var (
	DebugLogger *zap.Logger
	InfoLogger  *zap.Logger
	WarnLogger  *zap.Logger
	ErrorLogger *zap.Logger
	FatalLogger *zap.Logger
)

// DateWriter 按日期分割的日志写入器
type DateWriter struct {
	basePath    string
	currentDate string
	file        *os.File
	mu          sync.Mutex
}

// NewDateWriter 创建一个新的DateWriter
func NewDateWriter(basePath string) *DateWriter {
	return &DateWriter{
		basePath: basePath,
	}
}

// Write 实现io.Writer接口
func (w *DateWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	dateStr := now.Format("2006-01-02")

	// 如果日期改变，创建新文件
	if dateStr != w.currentDate {
		if w.file != nil {
			w.file.Close()
		}

		// 确保目录存在
		dir := filepath.Dir(w.basePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return 0, fmt.Errorf("failed to create log directory: %v", err)
		}

		// 创建新文件，文件名包含日期
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

// Close 关闭文件
func (w *DateWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// InitLogger 初始化日志系统
func InitLogger(level string, format string, outputPath string, maxSize int, maxBackups int, maxAge int, compress bool, rotateByDate bool, logInConsole bool) error {
	// 使用sync.Once确保日志系统只初始化一次
	once.Do(func() {
		initErr = initLoggerOnce(level, format, outputPath, maxSize, maxBackups, maxAge, compress, rotateByDate, logInConsole)
	})
	return initErr
}

// InitMultiLevelLogger 初始化多级别日志系统
func InitMultiLevelLogger(level string, format string, basePath string, maxSize int, maxBackups int, maxAge int, compress bool, logInConsole bool) error {
	once.Do(func() {
		initErr = initMultiLevelLoggerOnce(level, format, basePath, maxSize, maxBackups, maxAge, compress, logInConsole)
	})
	return initErr
}

// initLoggerOnce 实际执行日志初始化的函数
func initLoggerOnce(level string, format string, outputPath string, maxSize int, maxBackups int, maxAge int, compress bool, rotateByDate bool, logInConsole bool) error {
	// 解析日志级别
	zapLevel := parseLogLevel(level)

	// 创建编码器
	encoder := createEncoder(format)

	// 创建写入器
	writeSyncer := createWriteSyncer(outputPath, maxSize, maxBackups, maxAge, compress, rotateByDate, logInConsole)

	// 创建核心并初始化Logger
	core := zapcore.NewCore(encoder, writeSyncer, zapLevel)
	Logger = zap.New(core, zap.AddCaller())

	return nil
}

// parseLogLevel 解析日志级别
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

// createEncoder 创建日志编码器
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

// createWriteSyncer 创建日志写入器
func createWriteSyncer(outputPath string, maxSize int, maxBackups int, maxAge int, compress bool, rotateByDate bool, logInConsole bool) zapcore.WriteSyncer {
	// 基础文件日志写入器
	fileSyncer := createFileSyncer(outputPath, maxSize, maxBackups, maxAge, compress, rotateByDate)

	if logInConsole {
		// 同时输出到控制台和文件
		consoleSyncer := zapcore.AddSync(os.Stdout)
		return zapcore.NewMultiWriteSyncer(fileSyncer, consoleSyncer)
	}

	// 只输出到文件
	return fileSyncer
}

// createFileSyncer 创建文件同步器
func createFileSyncer(outputPath string, maxSize int, maxBackups int, maxAge int, compress bool, rotateByDate bool) zapcore.WriteSyncer {
	if rotateByDate {
		// 按日期分割
		dateWriter := NewDateWriter(outputPath)
		return zapcore.AddSync(dateWriter)
	}
	// 按大小分割
	return getLogWriter(outputPath, maxSize, maxBackups, maxAge, compress)
}

// getLogWriter 获取日志输出（lumberjack实现）
func getLogWriter(filePath string, maxSize int, maxBackups int, maxAge int, compress bool) zapcore.WriteSyncer {
	if filePath == "stdout" {
		return zapcore.AddSync(os.Stdout)
	}

	// 使用lumberjack实现日志切割
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   compress,
	}

	return zapcore.AddSync(lumberJackLogger)
}

// initMultiLevelLoggerOnce 初始化多级别日志系统
func initMultiLevelLoggerOnce(level string, format string, basePath string, maxSize int, maxBackups int, maxAge int, compress bool, logInConsole bool) error {
	// 获取当前日期
	currentDate := time.Now().Format("2006-01-02")

	// 创建日期目录
	logDir := filepath.Join(filepath.Dir(basePath), currentDate)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// 基础文件名
	baseName := filepath.Base(basePath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := baseName[:len(baseName)-len(ext)]

	// 为每个级别创建日志文件路径
	logFiles := map[string]string{
		"debug": filepath.Join(logDir, fmt.Sprintf("%s-debug%s", nameWithoutExt, ext)),
		"info":  filepath.Join(logDir, fmt.Sprintf("%s-info%s", nameWithoutExt, ext)),
		"warn":  filepath.Join(logDir, fmt.Sprintf("%s-warn%s", nameWithoutExt, ext)),
		"error": filepath.Join(logDir, fmt.Sprintf("%s-error%s", nameWithoutExt, ext)),
		"fatal": filepath.Join(logDir, fmt.Sprintf("%s-fatal%s", nameWithoutExt, ext)),
	}

	// 创建编码器
	encoder := createEncoder(format)

	// 为每个级别创建独立的logger
	logLevels := []string{"debug", "info", "warn", "error", "fatal"}
	loggers := []*zap.Logger{DebugLogger, InfoLogger, WarnLogger, ErrorLogger, FatalLogger}

	for i, level := range logLevels {
		writeSyncer := createLevelWriteSyncer(logFiles[level], maxSize, maxBackups, maxAge, compress, logInConsole)

		// 创建只包含当前级别及以上的核心
		levelCore := zapcore.NewCore(encoder, writeSyncer, parseLogLevel(level))

		// 如果是主logger，还需要包含所有级别
		if level == "info" {
			Logger = zap.New(levelCore, zap.AddCaller())
		}

		loggers[i] = zap.New(levelCore, zap.AddCaller())
	}

	// 更新全局变量
	DebugLogger = loggers[0]
	InfoLogger = loggers[1]
	WarnLogger = loggers[2]
	ErrorLogger = loggers[3]
	FatalLogger = loggers[4]

	return nil
}

// createLevelWriteSyncer 创建级别特定的日志写入器
func createLevelWriteSyncer(filePath string, maxSize int, maxBackups int, maxAge int, compress bool, logInConsole bool) zapcore.WriteSyncer {
	// 基础文件日志写入器
	fileSyncer := getLogWriter(filePath, maxSize, maxBackups, maxAge, compress)

	if logInConsole {
		// 同时输出到控制台和文件
		consoleSyncer := zapcore.AddSync(os.Stdout)
		return zapcore.NewMultiWriteSyncer(fileSyncer, consoleSyncer)
	}

	// 只输出到文件
	return fileSyncer
}

// Sync 同步日志缓冲区
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
	if DebugLogger != nil {
		DebugLogger.Sync()
	}
	if InfoLogger != nil {
		InfoLogger.Sync()
	}
	if WarnLogger != nil {
		WarnLogger.Sync()
	}
	if ErrorLogger != nil {
		ErrorLogger.Sync()
	}
	if FatalLogger != nil {
		FatalLogger.Sync()
	}
}

// Debug 记录调试日志
func Debug(msg string, fields ...zap.Field) {
	if DebugLogger != nil {
		DebugLogger.Debug(msg, fields...)
	}
}

// Info 记录信息日志
func Info(msg string, fields ...zap.Field) {
	if InfoLogger != nil {
		InfoLogger.Info(msg, fields...)
	}
}

// Warn 记录警告日志
func Warn(msg string, fields ...zap.Field) {
	if WarnLogger != nil {
		WarnLogger.Warn(msg, fields...)
	}
}

// Error 记录错误日志
func Error(msg string, fields ...zap.Field) {
	if ErrorLogger != nil {
		ErrorLogger.Error(msg, fields...)
	}
}

// Fatal 记录致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	if FatalLogger != nil {
		FatalLogger.Fatal(msg, fields...)
	}
}

// Uint 创建uint字段
func Uint(key string, val uint) zap.Field {
	return zap.Uint(key, val)
}

// Uint64 创建uint64字段
func Uint64(key string, val uint64) zap.Field {
	return zap.Uint64(key, val)
}

// Err 创建错误字段
func Err(err error) zap.Field {
	return zap.Error(err)
}

// String 创建字符串字段
func String(key string, val string) zap.Field {
	return zap.String(key, val)
}
