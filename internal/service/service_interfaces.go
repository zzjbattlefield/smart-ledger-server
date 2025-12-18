package service

import (
	"context"
	"mime/multipart"

	"smart-ledger-server/internal/model"
	"smart-ledger-server/internal/model/dto"
)

// 这里定义 handler 层依赖的最小服务接口，便于单测替身/Mock。

// UserServiceInterface 用户服务接口（供 Handler 依赖）
type UserServiceInterface interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.LoginResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	GetProfile(ctx context.Context, userID uint64) (*dto.UserResponse, error)
	UpdateProfile(ctx context.Context, userID uint64, req *dto.UpdateProfileRequest) (*dto.UserResponse, error)
}

// CategoryServiceInterface 分类服务接口（供 Handler 依赖）
type CategoryServiceInterface interface {
	Create(ctx context.Context, userID uint64, req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error)
	GetByID(ctx context.Context, userID, id uint64) (*dto.CategoryResponse, error)
	List(ctx context.Context, userID uint64, categoryType *int) ([]dto.CategoryResponse, error)
	Update(ctx context.Context, userID, id uint64, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error)
	Delete(ctx context.Context, userID, id uint64) error
	GetCategoriesForAI(ctx context.Context, userID uint64) ([]model.Category, error)
	InitFromTemplate(ctx context.Context, userID uint64) error
}

// BillServiceInterface 账单服务接口（供 Handler 依赖）
type BillServiceInterface interface {
	Create(ctx context.Context, userID uint64, req *dto.CreateBillRequest) (*dto.BillResponse, error)
	GetByID(ctx context.Context, userID, id uint64) (*dto.BillResponse, error)
	List(ctx context.Context, userID uint64, req *dto.BillListRequest) (*dto.BillListResponse, error)
	Update(ctx context.Context, userID, id uint64, req *dto.UpdateBillRequest) (*dto.BillResponse, error)
	Delete(ctx context.Context, userID, id uint64) error
	CreateFromAI(ctx context.Context, userID uint64, aiResult *dto.AIRecognizeResponse, imagePath string) (*dto.BillResponse, error)
	ImportFromExcel(ctx context.Context, userID uint64, filePath, parserType string) (*dto.BillImportResponse, error)
}

// StatsServiceInterface 统计服务接口（供 Handler 依赖）
type StatsServiceInterface interface {
	GetSummary(ctx context.Context, userID uint64, req *dto.StatsSummaryRequest) (*dto.StatsSummaryResponse, error)
	GetCategoryStats(ctx context.Context, userID uint64, req *dto.StatsCategoryRequest) (*dto.CategoryStatsResponse, error)
}

// AIServiceInterface AI服务接口（供 Handler 依赖）
type AIServiceInterface interface {
	RecognizeImage(ctx context.Context, userID uint64, file *multipart.FileHeader) (*dto.AIRecognizeResponse, error)
	RecognizeAndCreateBill(ctx context.Context, userID uint64, file *multipart.FileHeader) (*dto.BillResponse, error)
}
