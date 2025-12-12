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
	userRepo     *repository.UserRepository
	categoryRepo *repository.CategoryRepository
	billRepo     *repository.BillRepository

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

	// 初始化标志
	servicesInited bool
	handlersInited bool
}

// NewContainer 创建容器实例
func NewContainer(cfg *config.Config, db *gorm.DB, logger *zap.Logger) *Container {
	ctn := &Container{
		cfg:    cfg,
		db:     db,
		logger: logger,
	}

	// 立即初始化 Repositories
	ctn.userRepo = repository.NewUserRepository(db)
	ctn.categoryRepo = repository.NewCategoryRepository(db)
	ctn.billRepo = repository.NewBillRepository(db)

	return ctn
}

// Repository 访问器

// UserRepo 返回 UserRepository 实例
func (c *Container) UserRepo() *repository.UserRepository {
	return c.userRepo
}

// CategoryRepo 返回 CategoryRepository 实例
func (c *Container) CategoryRepo() *repository.CategoryRepository {
	return c.categoryRepo
}

// BillRepo 返回 BillRepository 实例
func (c *Container) BillRepo() *repository.BillRepository {
	return c.billRepo
}

// Service 访问器（惰性初始化）

// UserService 返回 UserService 实例
func (c *Container) UserService() *service.UserService {
	if !c.servicesInited {
		c.initServices()
	}
	return c.userService
}

// CategoryService 返回 CategoryService 实例
func (c *Container) CategoryService() *service.CategoryService {
	if !c.servicesInited {
		c.initServices()
	}
	return c.categoryService
}

// BillService 返回 BillService 实例
func (c *Container) BillService() *service.BillService {
	if !c.servicesInited {
		c.initServices()
	}
	return c.billService
}

// StatsService 返回 StatsService 实例
func (c *Container) StatsService() *service.StatsService {
	if !c.servicesInited {
		c.initServices()
	}
	return c.statsService
}

// AIService 返回 AIService 实例
func (c *Container) AIService() *service.AIService {
	if !c.servicesInited {
		c.initServices()
	}
	return c.aiService
}

// initServices 初始化所有 Services
func (c *Container) initServices() {
	if c.servicesInited {
		return
	}

	c.userService = service.NewUserService(c.userRepo, c.cfg)
	c.categoryService = service.NewCategoryService(c.categoryRepo)
	c.billService = service.NewBillService(c.billRepo, c.categoryRepo)
	c.statsService = service.NewStatsService(c.billRepo)

	// AI Service 可能失败
	aiService, err := service.NewAIService(&c.cfg.AI, c.billService, c.categoryService)
	if err != nil {
		c.logger.Warn("初始化AI服务失败", zap.Error(err))
	}
	c.aiService = aiService

	c.servicesInited = true
}

// Handler 访问器（惰性初始化）

// UserHandler 返回 UserHandler 实例
func (c *Container) UserHandler() *handler.UserHandler {
	if !c.handlersInited {
		c.initHandlers()
	}
	return c.userHandler
}

// CategoryHandler 返回 CategoryHandler 实例
func (c *Container) CategoryHandler() *handler.CategoryHandler {
	if !c.handlersInited {
		c.initHandlers()
	}
	return c.categoryHandler
}

// BillHandler 返回 BillHandler 实例
func (c *Container) BillHandler() *handler.BillHandler {
	if !c.handlersInited {
		c.initHandlers()
	}
	return c.billHandler
}

// StatsHandler 返回 StatsHandler 实例
func (c *Container) StatsHandler() *handler.StatsHandler {
	if !c.handlersInited {
		c.initHandlers()
	}
	return c.statsHandler
}

// AIHandler 返回 AIHandler 实例
func (c *Container) AIHandler() *handler.AIHandler {
	if !c.handlersInited {
		c.initHandlers()
	}
	return c.aiHandler
}

// initHandlers 初始化所有 Handlers
func (c *Container) initHandlers() {
	if c.handlersInited {
		return
	}

	// 确保 Services 已初始化
	if !c.servicesInited {
		c.initServices()
	}

	c.userHandler = handler.NewUserHandler(c.userService)
	c.categoryHandler = handler.NewCategoryHandler(c.categoryService)
	c.billHandler = handler.NewBillHandler(c.billService)
	c.statsHandler = handler.NewStatsHandler(c.statsService)

	if c.aiService != nil {
		c.aiHandler = handler.NewAIHandler(c.aiService)
	}

	c.handlersInited = true
}
