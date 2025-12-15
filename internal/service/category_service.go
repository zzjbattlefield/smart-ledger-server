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
	categoryRepo         CategoryRepo
	categoryTemplateRepo CategoryTemplateRepo
	// 缓存相关
	cacheMu sync.RWMutex
	cache   map[uint64]categoryCacheEntry
}

type categoryCacheEntry struct {
	categories []model.Category
	valid      bool
}

// NewCategoryService 创建分类服务
func NewCategoryService(categoryRepo CategoryRepo, categoryTemplateRepo CategoryTemplateRepo) *CategoryService {
	return &CategoryService{
		categoryRepo:         categoryRepo,
		categoryTemplateRepo: categoryTemplateRepo,
		cache:                make(map[uint64]categoryCacheEntry),
	}
}

// Create 创建分类
func (s *CategoryService) Create(ctx context.Context, userID uint64, req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
	// 检查名称是否重复
	exists, err := s.categoryRepo.ExistsByName(ctx, req.Name, userID, req.ParentID)
	if err != nil {
		return nil, errcode.ErrServer
	}
	if exists {
		return nil, errcode.ErrCategoryExists
	}

	// 如果有父分类，检查父分类是否存在
	if req.ParentID > 0 {
		parent, err := s.categoryRepo.GetByID(ctx, req.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errcode.ErrCategoryNotFound.WithMessage("父分类不存在")
			}
			return nil, errcode.ErrServer
		}
		if parent.UserID != userID {
			return nil, errcode.ErrCategoryNotFound.WithMessage("父分类不存在")
		}
	}

	category := &model.Category{
		Name:      req.Name,
		ParentID:  req.ParentID,
		UserID:    userID,
		Icon:      req.Icon,
		SortOrder: req.SortOrder,
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, errcode.ErrServer
	}

	// 使缓存失效
	s.invalidateCache(userID)

	return s.toCategoryResponse(category), nil
}

// GetByID 获取分类详情
func (s *CategoryService) GetByID(ctx context.Context, userID, id uint64) (*dto.CategoryResponse, error) {
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrCategoryNotFound
		}
		return nil, errcode.ErrServer
	}

	// 检查权限
	if category.UserID != userID {
		return nil, errcode.ErrForbidden
	}

	return s.toCategoryResponse(category), nil
}

// List 获取分类列表（树形结构）
func (s *CategoryService) List(ctx context.Context, userID uint64) ([]dto.CategoryResponse, error) {
	categories, err := s.categoryRepo.GetWithChildren(ctx, userID)
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
func (s *CategoryService) Update(ctx context.Context, userID, id uint64, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error) {
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrCategoryNotFound
		}
		return nil, errcode.ErrServer
	}

	// 检查权限
	if category.UserID != userID {
		return nil, errcode.ErrForbidden
	}

	// 检查名称是否重复
	if req.Name != "" && req.Name != category.Name {
		exists, err := s.categoryRepo.ExistsByName(ctx, req.Name, category.UserID, category.ParentID)
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
	s.invalidateCache(userID)

	return s.toCategoryResponse(category), nil
}

// Delete 删除分类
func (s *CategoryService) Delete(ctx context.Context, userID, id uint64) error {
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrCategoryNotFound
		}
		return errcode.ErrServer
	}

	// 检查权限
	if category.UserID != userID {
		return errcode.ErrForbidden
	}

	// 检查是否有子分类
	hasChildren, err := s.categoryRepo.HasChildren(ctx, userID, id)
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
	s.invalidateCache(userID)

	return nil
}

// GetByName 根据名称获取分类
func (s *CategoryService) GetByName(ctx context.Context, userID uint64, name string) (*model.Category, error) {
	return s.categoryRepo.GetByName(ctx, userID, name)
}

// toCategoryResponse 转换为分类响应
func (s *CategoryService) toCategoryResponse(category *model.Category) *dto.CategoryResponse {
	return &dto.CategoryResponse{
		ID:        category.ID,
		Name:      category.Name,
		ParentID:  category.ParentID,
		Icon:      category.Icon,
		SortOrder: category.SortOrder,
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
func (s *CategoryService) GetCategoriesForAI(ctx context.Context, userID uint64) ([]model.Category, error) {
	s.cacheMu.RLock()
	if entry, ok := s.cache[userID]; ok && entry.valid {
		categories := entry.categories
		s.cacheMu.RUnlock()
		return categories, nil
	}
	s.cacheMu.RUnlock()

	// 缓存未命中，重新加载
	return s.refreshCache(ctx, userID)
}

// refreshCache 刷新缓存
func (s *CategoryService) refreshCache(ctx context.Context, userID uint64) ([]model.Category, error) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	// 双重检查，避免重复加载
	if entry, ok := s.cache[userID]; ok && entry.valid {
		return entry.categories, nil
	}

	categories, err := s.categoryRepo.GetWithChildren(ctx, userID)
	if err != nil {
		return nil, err
	}

	s.cache[userID] = categoryCacheEntry{
		categories: categories,
		valid:      true,
	}
	return categories, nil
}

// invalidateCache 使缓存失效（按用户维度）
func (s *CategoryService) invalidateCache(userID uint64) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	delete(s.cache, userID)
}

// InitFromTemplate 从模板初始化用户分类
func (s *CategoryService) InitFromTemplate(ctx context.Context, userID uint64) error {
	templates, err := s.categoryTemplateRepo.GetAll(ctx)
	if err != nil {
		return err
	}
	idMap := make(map[uint64]uint64)
	for _, t := range templates {
		if t.ParentID == 0 {
			parentTemplate := model.Category{
				Name:      t.Name,
				ParentID:  0,
				UserID:    userID,
				Icon:      t.Icon,
				SortOrder: t.SortOrder,
			}
			if err := s.categoryRepo.Create(ctx, &parentTemplate); err != nil {
				return err
			}
			idMap[t.ID] = parentTemplate.ID
		}
	}
	for _, t := range templates {
		if t.ParentID != 0 {
			parentID := idMap[t.ParentID]
			childTemplate := model.Category{
				Name:      t.Name,
				ParentID:  parentID,
				UserID:    userID,
				Icon:      t.Icon,
				SortOrder: t.SortOrder,
			}
			if err := s.categoryRepo.Create(ctx, &childTemplate); err != nil {
				return err
			}
		}
	}
	s.invalidateCache(userID)
	return nil
}
