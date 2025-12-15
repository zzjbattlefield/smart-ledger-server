package repository

import (
	"context"

	"smart-ledger-server/internal/model"

	"gorm.io/gorm"
)

type CategoryTemplateRepository struct {
	db *gorm.DB
}

func NewCategoryTemplateRepository(db *gorm.DB) *CategoryTemplateRepository {
	return &CategoryTemplateRepository{db: db}
}

func (t *CategoryTemplateRepository) GetAll(ctx context.Context) ([]model.CategoryTemplate, error) {
	var templates []model.CategoryTemplate
	err := t.db.WithContext(ctx).Order("id ASC").Find(&templates).Error
	return templates, err
}
