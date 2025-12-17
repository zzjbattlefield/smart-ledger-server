package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"smart-ledger-server/internal/config"
	"smart-ledger-server/internal/container"
	"smart-ledger-server/internal/middleware"
)

// setupRouter 设置路由
func setupRouter(cfg *config.Config, ctn *container.Container) *gin.Engine {
	// 创建Gin实例
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// 注册路由
	registerRoutes(r, cfg, ctn)

	return r
}

// registerRoutes 注册路由
func registerRoutes(r *gin.Engine, cfg *config.Config, ctn *container.Container) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1
	v1 := r.Group("/v1")
	{
		registerUserPublicRoutes(v1, ctn)
		registerAuthenticatedRoutes(v1, cfg, ctn)
	}
}

// registerUserPublicRoutes 注册用户公开路由
func registerUserPublicRoutes(v1 *gin.RouterGroup, ctn *container.Container) {
	user := v1.Group("/user")
	h := ctn.UserHandler()
	{
		user.POST("/register", h.Register)
		user.POST("/login", h.Login)
	}
}

// registerAuthenticatedRoutes 注册需要认证的路由
func registerAuthenticatedRoutes(v1 *gin.RouterGroup, cfg *config.Config, ctn *container.Container) {
	auth := v1.Group("")
	auth.Use(middleware.Auth(&cfg.JWT, ctn.UserRepo()))
	{
		registerUserProtectedRoutes(auth, ctn)
		registerCategoryRoutes(auth, ctn)
		registerBillRoutes(auth, ctn)
		registerStatsRoutes(auth, ctn)
		registerAIRoutes(auth, ctn)
	}
}

// registerUserProtectedRoutes 注册用户受保护路由
func registerUserProtectedRoutes(auth *gin.RouterGroup, ctn *container.Container) {
	h := ctn.UserHandler()
	auth.GET("/user/profile", h.GetProfile)
	auth.PUT("/user/profile", h.UpdateProfile)
}

// registerCategoryRoutes 注册分类路由
func registerCategoryRoutes(auth *gin.RouterGroup, ctn *container.Container) {
	categories := auth.Group("/categories")
	h := ctn.CategoryHandler()
	{
		categories.GET("", h.List)
		categories.GET("/:id", h.Get)
		categories.POST("", h.Create)
		categories.PUT("/:id", h.Update)
		categories.DELETE("/:id", h.Delete)
	}
}

// registerBillRoutes 注册账单路由
func registerBillRoutes(auth *gin.RouterGroup, ctn *container.Container) {
	bills := auth.Group("/bills")
	h := ctn.BillHandler()
	{
		bills.GET("", h.List)
		bills.GET("/:id", h.Get)
		bills.POST("", h.Create)
		bills.POST("/import", h.Import)
		bills.PUT("/:id", h.Update)
		bills.DELETE("/:id", h.Delete)
	}
}

// registerStatsRoutes 注册统计路由
func registerStatsRoutes(auth *gin.RouterGroup, ctn *container.Container) {
	stats := auth.Group("/stats")
	h := ctn.StatsHandler()
	{
		stats.GET("/summary", h.GetSummary)
		stats.GET("/category", h.GetCategoryStats)
	}
}

// registerAIRoutes 注册AI路由
func registerAIRoutes(auth *gin.RouterGroup, ctn *container.Container) {
	if h := ctn.AIHandler(); h != nil {
		ai := auth.Group("/ai")
		{
			// 单张识别
			ai.POST("/recognize", h.Recognize)
			ai.POST("/recognize-and-save", h.RecognizeAndSave)
		}
	}
}
