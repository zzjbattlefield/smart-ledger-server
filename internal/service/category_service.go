package service

import (
	"context"
	"errors"
	"sync"

	"gorm.io/gorm"

	"smart-ledger-server/internal/model"
	"smart-ledger-server/internal/model/dto"
	"smart-ledger-server/pkg/errcode"
)

// CategoryService 分类服务
type CategoryService struct {
	categoryRepo CategoryRepo

	// 缓存相关
	cacheMu    sync.RWMutex
	cache      []model.Category
	cacheValid bool
}

// NewCategoryService 创建分类服务
func NewCategoryService(categoryRepo CategoryRepo) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
	}
}

// Create 创建分类
func (s *CategoryService) Create(ctx context.Context, req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
	// 检查名称是否重复
	exists, err := s.categoryRepo.ExistsByName(ctx, req.Name, req.ParentID, 0)
	if err != nil {
		return nil, errcode.ErrServer
	}
	if exists {
		return nil, errcode.ErrCategoryExists
	}

	// 如果有父分类，检查父分类是否存在
	if req.ParentID > 0 {
		_, err := s.categoryRepo.GetByID(ctx, req.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errcode.ErrCategoryNotFound.WithMessage("父分类不存在")
			}
			return nil, errcode.ErrServer
		}
	}

	category := &model.Category{
		Name:      req.Name,
		ParentID:  req.ParentID,
		Icon:      req.Icon,
		SortOrder: req.SortOrder,
		IsSystem:  false,
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, errcode.ErrServer
	}

	// 使缓存失效
	s.invalidateCache()

	return s.toCategoryResponse(category), nil
}

// GetByID 获取分类详情
func (s *CategoryService) GetByID(ctx context.Context, id uint64) (*dto.CategoryResponse, error) {
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrCategoryNotFound
		}
		return nil, errcode.ErrServer
	}

	return s.toCategoryResponse(category), nil
}

// List 获取分类列表（树形结构）
func (s *CategoryService) List(ctx context.Context) ([]dto.CategoryResponse, error) {
	categories, err := s.categoryRepo.GetWithChildren(ctx)
	if err != nil {
		return nil, errcode.ErrServer
	}

	result := make([]dto.CategoryResponse, len(categories))
	for i, cat := range categories {
		result[i] = *s.toCategoryResponseWithChildren(&cat)
	}

	return result, nil
}

// Update 更新分类
func (s *CategoryService) Update(ctx context.Context, id uint64, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error) {
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrCategoryNotFound
		}
		return nil, errcode.ErrServer
	}

	// 系统分类不可修改
	if category.IsSystem {
		return nil, errcode.ErrCategoryIsSystem
	}

	// 检查名称是否重复
	if req.Name != "" && req.Name != category.Name {
		exists, err := s.categoryRepo.ExistsByName(ctx, req.Name, category.ParentID, id)
		if err != nil {
			return nil, errcode.ErrServer
		}
		if exists {
			return nil, errcode.ErrCategoryExists
		}
		category.Name = req.Name
	}

	if req.Icon != "" {
		category.Icon = req.Icon
	}
	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, errcode.ErrServer
	}

	// 使缓存失效
	s.invalidateCache()

	return s.toCategoryResponse(category), nil
}

// Delete 删除分类
func (s *CategoryService) Delete(ctx context.Context, id uint64) error {
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrCategoryNotFound
		}
		return errcode.ErrServer
	}

	// 系统分类不可删除
	if category.IsSystem {
		return errcode.ErrCategoryIsSystem
	}

	// 检查是否有子分类
	hasChildren, err := s.categoryRepo.HasChildren(ctx, id)
	if err != nil {
		return errcode.ErrServer
	}
	if hasChildren {
		return errcode.ErrCategoryHasChildren
	}

	if err := s.categoryRepo.Delete(ctx, id); err != nil {
		return errcode.ErrServer
	}

	// 使缓存失效
	s.invalidateCache()

	return nil
}

// GetByName 根据名称获取分类
func (s *CategoryService) GetByName(ctx context.Context, name string) (*model.Category, error) {
	return s.categoryRepo.GetByName(ctx, name)
}

// toCategoryResponse 转换为分类响应
func (s *CategoryService) toCategoryResponse(category *model.Category) *dto.CategoryResponse {
	return &dto.CategoryResponse{
		ID:        category.ID,
		Name:      category.Name,
		ParentID:  category.ParentID,
		Icon:      category.Icon,
		SortOrder: category.SortOrder,
		IsSystem:  category.IsSystem,
	}
}

// toCategoryResponseWithChildren 转换为分类响应（含子分类）
func (s *CategoryService) toCategoryResponseWithChildren(category *model.Category) *dto.CategoryResponse {
	resp := s.toCategoryResponse(category)
	if len(category.Children) > 0 {
		resp.Children = make([]dto.CategoryResponse, len(category.Children))
		for i, child := range category.Children {
			resp.Children[i] = *s.toCategoryResponse(&child)
		}
	}
	return resp
}

// GetCategoriesForAI 获取分类数据（带缓存，供 AI 识别使用）
func (s *CategoryService) GetCategoriesForAI(ctx context.Context) ([]model.Category, error) {
	s.cacheMu.RLock()
	if s.cacheValid {
		defer s.cacheMu.RUnlock()
		return s.cache, nil
	}
	s.cacheMu.RUnlock()

	// 缓存未命中，重新加载
	return s.refreshCache(ctx)
}

// refreshCache 刷新缓存
func (s *CategoryService) refreshCache(ctx context.Context) ([]model.Category, error) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	// 双重检查，避免重复加载
	if s.cacheValid {
		return s.cache, nil
	}

	categories, err := s.categoryRepo.GetWithChildren(ctx)
	if err != nil {
		return nil, err
	}

	s.cache = categories
	s.cacheValid = true
	return categories, nil
}

// invalidateCache 使缓存失效
func (s *CategoryService) invalidateCache() {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	s.cacheValid = false
}
