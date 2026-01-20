package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mygoframe/pkg/cache"
	"mygoframe/pkg/config"
	"mygoframe/pkg/database"
	"mygoframe/pkg/logger"
	"mygoframe/routes"

	"go.uber.org/zap"
)

func main() {
	Run()
}

func Run() {
	cfg := config.GetConfig()

	if err := logger.InitLogger(cfg.Zap); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	// 初始化缓存系统
	if err := cache.Init(cfg); err != nil {
		logger.Error("初始化缓存失败", zap.Error(err))
	}

	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	r := routes.SetupRoutes(db)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.System.Addr),
		Handler: r,
	}

	go func() {
		log.Printf("服务器启动，监听端口: %d", cfg.System.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("服务器关闭失败: %v", err)
	}

	if err := cache.Close(); err != nil {
		logger.Error("关闭缓存系统失败", zap.Error(err))
	}

	log.Println("服务器已关闭")
}
