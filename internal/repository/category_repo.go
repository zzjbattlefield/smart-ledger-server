package repository

import (
	"context"

	"gorm.io/gorm"

	"smart-ledger-server/internal/model"
)

// CategoryRepository 分类数据访问层
type CategoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository 创建分类仓库
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Create 创建分类
func (r *CategoryRepository) Create(ctx context.Context, category *model.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

// GetByID 根据ID获取分类
func (r *CategoryRepository) GetByID(ctx context.Context, id uint64) (*model.Category, error) {
	var category model.Category
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// GetAll 获取所有分类
func (r *CategoryRepository) GetAll(ctx context.Context, userID uint64) ([]model.Category, error) {
	var categories []model.Category
	err := r.db.WithContext(ctx).Where("user_id", userID).Order("sort_order ASC, id ASC").Find(&categories).Error
	return categories, err
}

// GetByParentID 根据父ID获取分类
func (r *CategoryRepository) GetByParentID(ctx context.Context, parentID uint64) ([]model.Category, error) {
	var categories []model.Category
	err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Order("sort_order ASC, id ASC").Find(&categories).Error
	return categories, err
}

// GetTopLevel 获取顶级分类
func (r *CategoryRepository) GetTopLevel(ctx context.Context) ([]model.Category, error) {
	return r.GetByParentID(ctx, 0)
}

// GetWithChildren 获取分类及其子分类
func (r *CategoryRepository) GetWithChildren(ctx context.Context, userID uint64) ([]model.Category, error) {
	var categories []model.Category
	err := r.db.WithContext(ctx).
		Where("parent_id = 0 and user_id = ?", userID).
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Where("user_id = ?", userID).Order("sort_order ASC, id ASC")
		}).
		Order("sort_order ASC, id ASC").
		Find(&categories).Error
	return categories, err
}

// Update 更新分类
func (r *CategoryRepository) Update(ctx context.Context, category *model.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

// Delete 删除分类
func (r *CategoryRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Category{}, id).Error
}

// HasChildren 检查是否有子分类
func (r *CategoryRepository) HasChildren(ctx context.Context, userID, id uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Category{}).Where("parent_id = ? AND user_id = ?", id, userID).Count(&count).Error
	return count > 0, err
}

// ExistsByName 检查分类名是否存在(同一父级下)
func (r *CategoryRepository) ExistsByName(ctx context.Context, name string, userID, parentID uint64) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.Category{}).Where("name = ? AND user_id = ? AND parent_id = ?", name, userID, parentID)
	err := query.Count(&count).Error
	return count > 0, err
}

// GetByName 根据名称获取分类
func (r *CategoryRepository) GetByName(ctx context.Context, userID uint64, name string) (*model.Category, error) {
	var category model.Category
	err := r.db.WithContext(ctx).Where("name = ? AND user_id = ?", name, userID).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}
