package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fmt"

	"github.com/ouyuan2016/mygoframe/routes"
	"go.uber.org/zap"

	"github.com/ouyuan2016/mygoframe/pkg/config"
	"github.com/ouyuan2016/mygoframe/pkg/database"
	"github.com/ouyuan2016/mygoframe/pkg/logger"
)

func Run() error {
	// 加载配置
	cfg := config.GetConfig()

	// 初始化日志系统（使用新的多级别日志系统）
	err := logger.InitMultiLevelLogger(
		cfg.Zap.Level,
		cfg.Zap.Format,
		cfg.Zap.OutputPath,
		cfg.Zap.MaxSize,
		cfg.Zap.MaxBackups,
		cfg.Zap.MaxAge,
		cfg.Zap.Compress,
		cfg.Zap.LogInConsole,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// 记录启动日志
	logger.Info("Starting StarLive server...")

	// 初始化数据库连接
	err = database.InitDB(cfg)
	if err != nil {
		logger.Error("Failed to initialize database", zap.Error(err))
		return err
	}
	logger.Info("Database initialized successfully")

	// 设置路由
	r := routes.SetupRoutes()

	// 创建HTTP服务器
	port := fmt.Sprintf(":%d", cfg.System.Addr)
	srv := &http.Server{
		Addr:    port,
		Handler: r,
	}

	// 在goroutine中启动服务器
	go func() {
		logger.Info("Server starting", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", zap.Error(err))
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// 设置一个超时时间，确保服务器在合理时间内关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭HTTP服务器
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
		return err
	}

	logger.Info("Server exited")
	return nil
}
