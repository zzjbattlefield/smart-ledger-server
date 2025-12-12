package service

import (
	"context"
	"time"

	"smart-ledger-server/internal/model"
	"smart-ledger-server/internal/repository"
)

// 这里定义 service 层依赖的最小仓库接口（依赖反转）。
// 仅暴露当前业务真正使用的方法，方便单测替身/Mock。

// UserRepo 用户仓库接口
type UserRepo interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uint64) (*model.User, error)
	GetByPhone(ctx context.Context, phone string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	UpdateFields(ctx context.Context, id uint64, fields map[string]interface{}) error
	ExistsByPhone(ctx context.Context, phone string) (bool, error)
}

// CategoryRepo 分类仓库接口
type CategoryRepo interface {
	Create(ctx context.Context, category *model.Category) error
	GetByID(ctx context.Context, id uint64) (*model.Category, error)
	GetWithChildren(ctx context.Context) ([]model.Category, error)
	Update(ctx context.Context, category *model.Category) error
	Delete(ctx context.Context, id uint64) error
	HasChildren(ctx context.Context, id uint64) (bool, error)
	ExistsByName(ctx context.Context, name string, parentID uint64, excludeID uint64) (bool, error)
	GetByName(ctx context.Context, name string) (*model.Category, error)
}

// BillRepo 账单仓库接口
type BillRepo interface {
	CreateWithItems(ctx context.Context, bill *model.Bill, items []model.BillItem) error
	GetByID(ctx context.Context, id uint64) (*model.Bill, error)
	List(ctx context.Context, query *repository.BillQuery) ([]model.Bill, int64, error)
	Update(ctx context.Context, bill *model.Bill) error
	Delete(ctx context.Context, id uint64) error

	// 统计相关
	GetStatsSummary(ctx context.Context, userID uint64, startDate, endDate time.Time) (*repository.StatsSummary, error)
	GetCategoryStats(ctx context.Context, userID uint64, billType model.BillType, startDate, endDate time.Time) ([]repository.CategoryStats, error)
	GetDailyStats(ctx context.Context, userID uint64, startDate, endDate time.Time) ([]repository.DailyStats, error)
}
