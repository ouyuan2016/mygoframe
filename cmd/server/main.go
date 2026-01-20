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

	"mygoframe/internal/task"
	"mygoframe/pkg/cache"
	"mygoframe/pkg/config"
	"mygoframe/pkg/database"
	"mygoframe/pkg/logger"
	"mygoframe/pkg/queue"
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

	// 初始化队列
	if cfg.Queue.Enabled {
		queue.InitQueue()
		queue.InitQueueServer()
		task.Setup() // 设置和注册所有任务

		// 注册处理器到Mux
		for pattern, handler := range queue.HandleFunc {
			h := handler
			queue.Mux.HandleFunc(pattern, h)
		}

		go func() {
			log.Println("队列服务器启动")
			if err := queue.Server.Run(queue.Mux); err != nil {
				log.Fatalf("队列服务器启动失败: %v", err)
			}
		}()
		go func() {
			log.Println("定时任务调度器启动")
			if err := queue.Scheduler.Run(); err != nil {
				log.Fatalf("定时任务调度器启动失败: %v", err)
			}
		}()
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

	if cfg.Queue.Enabled {
		queue.StopScheduler()
		log.Println("定时任务调度器已关闭")
		queue.Server.Shutdown()
		log.Println("队列服务器已关闭")
	}

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
