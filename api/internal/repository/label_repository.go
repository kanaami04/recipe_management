package repository

import (
	"context"

	"recipe-backend/internal/domain"

	"gorm.io/gorm"
)

type labelRepository struct {
	db *gorm.DB
}

func NewLabelRepository(db *gorm.DB) domain.LabelRepository {
	return &labelRepository{db: db}
}

func (r *labelRepository) FindAll(ctx context.Context) ([]domain.RecipeLabel, error) {
	var labels []domain.RecipeLabel
	err := r.db.WithContext(ctx).Order("id").Find(&labels).Error
	return labels, err
}
