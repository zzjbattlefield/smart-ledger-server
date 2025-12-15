package container

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"smart-ledger-server/internal/config"
	"smart-ledger-server/internal/handler"
	"smart-ledger-server/internal/repository"
	"smart-ledger-server/internal/service"
)

// Container 依赖注入容器，封装和管理所有依赖
type Container struct {
	// 基础设施
	cfg    *config.Config
	db     *gorm.DB
	logger *zap.Logger

	// Repositories
	userRepo             *repository.UserRepository
	categoryRepo         *repository.CategoryRepository
	billRepo             *repository.BillRepository
	categoryTemplateRepo *repository.CategoryTemplateRepository

	// Services
	userService     *service.UserService
	categoryService *service.CategoryService
	billService     *service.BillService
	statsService    *service.StatsService
	aiService       *service.AIService

	// Handlers
	userHandler     *handler.UserHandler
	categoryHandler *handler.CategoryHandler
	billHandler     *handler.BillHandler
	statsHandler    *handler.StatsHandler
	aiHandler       *handler.AIHandler
}

// NewContainer 创建容器实例
func NewContainer(cfg *config.Config, db *gorm.DB, logger *zap.Logger) *Container {
	ctn := &Container{
		cfg:    cfg,
		db:     db,
		logger: logger,
	}
	ctn.initRepositories()
	ctn.initServices()
	ctn.initHandlers()
	return ctn
}

// initRepositories 初始化所有 Repositories
func (c *Container) initRepositories() {
	c.userRepo = repository.NewUserRepository(c.db)
	c.categoryRepo = repository.NewCategoryRepository(c.db)
	c.billRepo = repository.NewBillRepository(c.db)
	c.categoryTemplateRepo = repository.NewCategoryTemplateRepository(c.db)
}

// initServices 初始化所有 Services
func (c *Container) initServices() {
	c.categoryService = service.NewCategoryService(c.categoryRepo, c.categoryTemplateRepo)
	c.userService = service.NewUserService(c.userRepo, c.categoryService, c.cfg)
	c.billService = service.NewBillService(c.billRepo, c.categoryRepo)
	c.statsService = service.NewStatsService(c.billRepo)

	// AI Service 可能失败
	aiService, err := service.NewAIService(&c.cfg.AI, c.billService, c.categoryService)
	if err != nil {
		c.logger.Warn("初始化AI服务失败", zap.Error(err))
	}
	c.aiService = aiService
}

// initHandlers 初始化所有 Handlers
func (c *Container) initHandlers() {
	c.userHandler = handler.NewUserHandler(c.userService)
	c.categoryHandler = handler.NewCategoryHandler(c.categoryService)
	c.billHandler = handler.NewBillHandler(c.billService)
	c.statsHandler = handler.NewStatsHandler(c.statsService)
	if c.aiService != nil {
		c.aiHandler = handler.NewAIHandler(c.aiService)
	}
}

// Repository 访问器

func (c *Container) UserRepo() *repository.UserRepository         { return c.userRepo }
func (c *Container) CategoryRepo() *repository.CategoryRepository { return c.categoryRepo }
func (c *Container) BillRepo() *repository.BillRepository         { return c.billRepo }

// Service 访问器

func (c *Container) UserService() *service.UserService         { return c.userService }
func (c *Container) CategoryService() *service.CategoryService { return c.categoryService }
func (c *Container) BillService() *service.BillService         { return c.billService }
func (c *Container) StatsService() *service.StatsService       { return c.statsService }
func (c *Container) AIService() *service.AIService             { return c.aiService }

// Handler 访问器

func (c *Container) UserHandler() *handler.UserHandler         { return c.userHandler }
func (c *Container) CategoryHandler() *handler.CategoryHandler { return c.categoryHandler }
func (c *Container) BillHandler() *handler.BillHandler         { return c.billHandler }
func (c *Container) StatsHandler() *handler.StatsHandler       { return c.statsHandler }
func (c *Container) AIHandler() *handler.AIHandler             { return c.aiHandler }
