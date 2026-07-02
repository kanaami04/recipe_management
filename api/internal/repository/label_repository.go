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

func (r *labelRepository) FindNamesForUser(ctx context.Context, userID string) ([]string, error) {
	var names []string
	db := r.db.WithContext(ctx)
	err := db.Model(&domain.RecipeLabel{}).
		Distinct().
		Joins("JOIN recipes ON recipes.id = recipe_labels.recipe_id").
		Where("recipes.owner_id = ? OR recipes.id IN (?)", userID, sharedRecipeIDs(db, userID)).
		Order("recipe_labels.name").
		Pluck("recipe_labels.name", &names).Error
	return names, err
}
