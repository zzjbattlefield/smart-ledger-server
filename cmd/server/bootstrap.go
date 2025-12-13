package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"smart-ledger-server/internal/config"
	"smart-ledger-server/internal/middleware"
	"smart-ledger-server/internal/pkg/database"
	"smart-ledger-server/internal/pkg/logger"
)

// mustLoadConfig 加载配置，失败则退出
func mustLoadConfig(path string) *config.Config {
	cfg, err := config.Load(path)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}
	return cfg
}

// mustInitLogger 初始化日志，失败则退出
func mustInitLogger(cfg *config.LogConfig) *zap.Logger {
	log, err := logger.Init(cfg)
	if err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	return log
}

// mustInitDatabase 初始化数据库，失败则退出
func mustInitDatabase(cfg *config.DatabaseConfig, log *zap.Logger) *gorm.DB {
	if err := database.Init(cfg, log); err != nil {
		log.Fatal("初始化数据库失败", zap.Error(err))
	}
	return database.GetDB()
}

// mustMigrateDatabase 执行数据库迁移，失败则退出
func mustMigrateDatabase(log *zap.Logger) {
	if err := database.AutoMigrate(); err != nil {
		log.Fatal("数据库迁移失败", zap.Error(err))
	}
	log.Info("数据库迁移完成")
}

// initRedisIfConfigured 如果配置了 Redis 则初始化
// 返回 true 表示 Redis 初始化成功，调用者需要在程序退出时调用 database.CloseRedis()
func initRedisIfConfigured(cfg *config.RedisConfig, log *zap.Logger) bool {
	if cfg.Host != "" {
		if err := database.InitRedis(cfg, log); err != nil {
			log.Warn("初始化Redis失败，将不使用缓存", zap.Error(err))
			return false
		}
		return true
	}
	return false
}

// applyGlobalMiddleware 应用全局中间件
func applyGlobalMiddleware(r *gin.Engine, log *zap.Logger, limiter *middleware.IPRateLimiter) {
	r.Use(middleware.Recovery(log))
	r.Use(middleware.Logger(log))
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(limiter))
}

// runServer 启动服务器并处理优雅关闭
func runServer(cfg *config.Config, r *gin.Engine, log *zap.Logger) {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 启动服务器
	go func() {
		log.Info("服务器启动", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("服务器启动失败", zap.Error(err))
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("正在关闭服务器...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("服务器强制关闭", zap.Error(err))
	}

	log.Info("服务器已关闭")
}
