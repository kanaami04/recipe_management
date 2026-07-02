package repository

import (
	"context"
	"errors"

	"recipe-backend/internal/domain"

	"gorm.io/gorm"
)

type recipeRepository struct {
	db *gorm.DB
}

func NewRecipeRepository(db *gorm.DB) domain.RecipeRepository {
	return &recipeRepository{db: db}
}

// preloadAll は関連を全て eager load するためのスコープ。
func preloadAll(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Owner").
		Preload("Labels").
		Preload("SharedUsers").
		Preload("Ingredients").
		Preload("Seasonings")
}

// sharedRecipeIDs は共有先に userID を含むレシピ ID のサブクエリを返す。
func sharedRecipeIDs(db *gorm.DB, userID uint) *gorm.DB {
	return db.Table("recipe_shares").
		Select("recipe_id").
		Where("user_id = ?", userID)
}

func (r *recipeRepository) FindAllForUser(ctx context.Context, userID uint) ([]domain.Recipe, error) {
	var recipes []domain.Recipe
	db := r.db.WithContext(ctx)
	err := preloadAll(db).
		Where("owner_id = ?", userID).
		Or("id IN (?)", sharedRecipeIDs(db, userID)).
		Order("id").
		Find(&recipes).Error
	return recipes, err
}

func (r *recipeRepository) FindByID(ctx context.Context, id uint) (*domain.Recipe, error) {
	var recipe domain.Recipe
	err := preloadAll(r.db.WithContext(ctx)).First(&recipe, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &recipe, nil
}

func (r *recipeRepository) Create(ctx context.Context, recipe *domain.Recipe) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 子テーブル(Labels/Ingredients/Seasonings)は本体と一緒に insert される。
		// SharedUsers(m2m)は既存ユーザーへの参照のため Replace で中間テーブルのみ操作する。
		if err := tx.Omit("Owner", "SharedUsers").Create(recipe).Error; err != nil {
			return err
		}
		return tx.Model(recipe).Association("SharedUsers").Replace(recipe.SharedUsers)
	})
}

func (r *recipeRepository) Update(ctx context.Context, recipe *domain.Recipe) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// スカラー項目のみ更新（owner / created_at は変更しない）。
		// cooking_time は NULL を書ける必要があるため map で更新する。
		if err := tx.Model(&domain.Recipe{ID: recipe.ID}).Updates(map[string]interface{}{
			"title":        recipe.Title,
			"cooking_time": recipe.CookingTime,
			"servings":     recipe.Servings,
			"procedure":    recipe.Procedure,
			"archived":     recipe.Archived,
		}).Error; err != nil {
			return err
		}

		// 子テーブルは全置換（差分更新はしない）。
		if err := replaceChildren(tx, recipe); err != nil {
			return err
		}
		return tx.Model(recipe).Association("SharedUsers").Replace(recipe.SharedUsers)
	})
}

// replaceChildren はレシピ従属の子テーブル(材料・調味料・ラベル)を削除して作り直す。
func replaceChildren(tx *gorm.DB, recipe *domain.Recipe) error {
	if err := tx.Where("recipe_id = ?", recipe.ID).Delete(&domain.RecipeIngredient{}).Error; err != nil {
		return err
	}
	if err := tx.Where("recipe_id = ?", recipe.ID).Delete(&domain.RecipeSeasoning{}).Error; err != nil {
		return err
	}
	if err := tx.Where("recipe_id = ?", recipe.ID).Delete(&domain.RecipeLabel{}).Error; err != nil {
		return err
	}

	if len(recipe.Ingredients) > 0 {
		for i := range recipe.Ingredients {
			recipe.Ingredients[i].ID = 0
			recipe.Ingredients[i].RecipeID = recipe.ID
		}
		if err := tx.Create(&recipe.Ingredients).Error; err != nil {
			return err
		}
	}
	if len(recipe.Seasonings) > 0 {
		for i := range recipe.Seasonings {
			recipe.Seasonings[i].ID = 0
			recipe.Seasonings[i].RecipeID = recipe.ID
		}
		if err := tx.Create(&recipe.Seasonings).Error; err != nil {
			return err
		}
	}
	if len(recipe.Labels) > 0 {
		for i := range recipe.Labels {
			recipe.Labels[i].ID = 0
			recipe.Labels[i].RecipeID = recipe.ID
		}
		if err := tx.Create(&recipe.Labels).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *recipeRepository) Delete(ctx context.Context, recipe *domain.Recipe) error {
	// 子テーブル・中間テーブル(recipe_shares)は FK の ON DELETE CASCADE で削除される。
	return r.db.WithContext(ctx).Delete(recipe).Error
}
