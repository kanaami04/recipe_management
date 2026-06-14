package repository

import (
	"recipe-backend/internal/domain"

	"gorm.io/gorm"
)

// Repositories は全リポジトリを束ねる。
type Repositories struct {
	User   domain.UserRepository
	Label  domain.LabelRepository
	Recipe domain.RecipeRepository
}

// New は db から全リポジトリを生成する（リポジトリ層の合成）。
func New(db *gorm.DB) *Repositories {
	return &Repositories{
		User:   NewUserRepository(db),
		Label:  NewLabelRepository(db),
		Recipe: NewRecipeRepository(db),
	}
}
