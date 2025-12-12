package main

import (
	"flag"

	"golang.org/x/time/rate"

	"smart-ledger-server/internal/container"
	"smart-ledger-server/internal/middleware"
	"smart-ledger-server/internal/pkg/database"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "configs/config.yaml", "配置文件路径")
}

func main() {
	flag.Parse()

	// 1. 加载配置
	cfg := mustLoadConfig(configPath)

	// 2. 初始化基础设施
	log := mustInitLogger(&cfg.Log)
	defer log.Sync()

	db := mustInitDatabase(&cfg.Database, log)
	defer database.Close()

	mustMigrateDatabase(log)

	// 初始化 Redis（如果配置了）
	if initRedisIfConfigured(&cfg.Redis, log) {
		defer database.CloseRedis()
	}

	// 3. 创建依赖容器（核心简化点）
	ctn := container.NewContainer(cfg, db, log)

	// 4. 初始化限流器
	ipLimiter := middleware.NewIPRateLimiter(rate.Limit(10), 20)

	// 5. 设置路由（只传递容器）
	r := setupRouter(cfg, ctn)

	// 6. 应用全局中间件
	applyGlobalMiddleware(r, log, ipLimiter)

	// 7. 启动服务器
	runServer(cfg, r, log)
}
