package main

import (
	"mygoframe/pkg/config"
	"mygoframe/pkg/logger"
	"time"

	"go.uber.org/zap"
)

func main() {
	// 初始化配置
	cfg := config.Zap{
		Level:           "debug",            // 设置为 debug 级别，显示所有日志
		Format:          "console",          // 控制台格式，更易读
		OutputPath:      "log/starlive.log", // 输出路径
		RotateByDate:    true,               // 按日期分割
		LevelSeparation: true,               // 按级别分离文件
		LogInConsole:    true,               // 同时在控制台输出
		MaxSize:         100,                // 最大文件大小 (MB)
		MaxBackups:      3,                  // 最大备份数量
		MaxAge:          7,                  // 最大保存天数
		Compress:        false,              // 不压缩
	}

	// 初始化日志器
	if err := logger.InitLogger(cfg); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	// 测试不同级别的日志
	logger.Debug("这是一个调试日志", zap.String("module", "test"), zap.Int("count", 1))
	logger.Info("这是一个信息日志", zap.String("user", "admin"), zap.String("action", "login"))
	logger.Warn("这是一个警告日志", zap.String("reason", "high_memory"), zap.Float64("usage", 85.5))
	logger.Error("这是一个错误日志", zap.String("error_code", "DB_CONNECTION_FAILED"), zap.String("database", "mysql"))

	// 测试带多个字段的日志
	logger.Info("用户操作日志",
		zap.String("user_id", "12345"),
		zap.String("ip", "192.168.1.100"),
		zap.String("action", "create_post"),
		zap.Int("post_id", 67890),
		zap.Duration("duration", 150*time.Millisecond),
	)

	// 测试错误链
	err := &TestError{code: "VALIDATION_FAILED", message: "参数验证失败"}
	logger.Error("业务处理失败", logger.Err(err), zap.String("request_id", "req_001"))

	// 测试批量日志
	for i := 0; i < 5; i++ {
		logger.Info("批量日志测试",
			zap.Int("index", i),
			zap.String("status", "processing"),
			zap.Time("timestamp", time.Now()),
		)
		time.Sleep(100 * time.Millisecond) // 稍微延迟，模拟真实场景
	}

	// 测试不同级别的并发日志
	go func() {
		for i := 0; i < 3; i++ {
			logger.Debug("并发调试日志", zap.Int("goroutine", 1), zap.Int("count", i))
			time.Sleep(50 * time.Millisecond)
		}
	}()

	go func() {
		for i := 0; i < 3; i++ {
			logger.Info("并发信息日志", zap.Int("goroutine", 2), zap.String("task", "background_job"))
			time.Sleep(60 * time.Millisecond)
		}
	}()

	// 等待并发日志完成
	time.Sleep(300 * time.Millisecond)

	logger.Info("日志测试完成！", zap.String("result", "success"))
	logger.Info("请检查 log 目录下的文件结构")
	logger.Info("传统模式文件: log/starlive-YYYY-MM-DD.log")
	logger.Info("新模式目录结构: log/YYYY-MM-DD/{debug,info,warn,error}.log")
}

// TestError 自定义错误类型
type TestError struct {
	code    string
	message string
}

func (e *TestError) Error() string {
	return e.message
}

func (e *TestError) Code() string {
	return e.code
}
