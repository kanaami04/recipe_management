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
		Preload("Cooking.Ingredient").
		Preload("Season.Seasoning")
}

func (r *recipeRepository) FindAllForUser(ctx context.Context, userID uint) ([]domain.Recipe, error) {
	var recipes []domain.Recipe
	db := r.db.WithContext(ctx)
	// 共有先に userID を含むレシピIDのサブクエリ
	sub := db.Table("recipes_shared_user").
		Select("recipe_id").
		Where("application_user_id = ?", userID)

	err := preloadAll(db).
		Where("owner_id = ?", userID).
		Or("id IN (?)", sub).
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
		labels := recipe.Labels
		shared := recipe.SharedUsers
		cooking := recipe.Cooking
		season := recipe.Season

		// 関連は手動制御するため、本体のみ作成
		if err := tx.Omit("Labels", "SharedUsers", "Cooking", "Season").Create(recipe).Error; err != nil {
			return err
		}
		return r.saveAssociations(tx, recipe, labels, shared, cooking, season)
	})
}

func (r *recipeRepository) Update(ctx context.Context, recipe *domain.Recipe) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		labels := recipe.Labels
		shared := recipe.SharedUsers
		cooking := recipe.Cooking
		season := recipe.Season

		// スカラー項目のみ更新（owner / created_at は変更しない）。
		// create_time は NULL を書ける必要があるため map で更新する。
		if err := tx.Model(&domain.Recipe{ID: recipe.ID}).Updates(map[string]interface{}{
			"title":       recipe.Title,
			"create_time": recipe.CreateTime,
			"create_for":  recipe.CreateFor,
			"procedure":   recipe.Procedure,
			"archive_flg": recipe.ArchiveFlg,
		}).Error; err != nil {
			return err
		}

		// 子テーブル（cooking/season）は置換のため一旦削除
		if err := tx.Where("recipe_id = ?", recipe.ID).Delete(&domain.Cooking{}).Error; err != nil {
			return err
		}
		if err := tx.Where("recipe_id = ?", recipe.ID).Delete(&domain.Season{}).Error; err != nil {
			return err
		}
		return r.saveAssociations(tx, recipe, labels, shared, cooking, season)
	})
}

// saveAssociations は m2m（label/shared_user）を Replace し、cooking/season を作成する。
func (r *recipeRepository) saveAssociations(
	tx *gorm.DB,
	recipe *domain.Recipe,
	labels []domain.RecipeLabel,
	shared []domain.ApplicationUser,
	cooking []domain.Cooking,
	season []domain.Season,
) error {
	if err := tx.Model(recipe).Association("Labels").Replace(labels); err != nil {
		return err
	}
	if err := tx.Model(recipe).Association("SharedUsers").Replace(shared); err != nil {
		return err
	}
	if len(cooking) > 0 {
		for i := range cooking {
			cooking[i].RecipeID = recipe.ID
		}
		if err := tx.Omit("Ingredient").Create(&cooking).Error; err != nil {
			return err
		}
	}
	if len(season) > 0 {
		for i := range season {
			season[i].RecipeID = recipe.ID
		}
		if err := tx.Omit("Seasoning").Create(&season).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *recipeRepository) Delete(ctx context.Context, recipe *domain.Recipe) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("recipe_id = ?", recipe.ID).Delete(&domain.Cooking{}).Error; err != nil {
			return err
		}
		if err := tx.Where("recipe_id = ?", recipe.ID).Delete(&domain.Season{}).Error; err != nil {
			return err
		}
		if err := tx.Model(recipe).Association("Labels").Clear(); err != nil {
			return err
		}
		if err := tx.Model(recipe).Association("SharedUsers").Clear(); err != nil {
			return err
		}
		return tx.Delete(recipe).Error
	})
}

func (r *recipeRepository) GetOrCreateLabel(ctx context.Context, name string) (*domain.RecipeLabel, error) {
	var label domain.RecipeLabel
	err := r.db.WithContext(ctx).FirstOrCreate(&label, domain.RecipeLabel{Name: name}).Error
	return &label, err
}

func (r *recipeRepository) GetOrCreateIngredient(ctx context.Context, name string) (*domain.Ingredient, error) {
	var ing domain.Ingredient
	err := r.db.WithContext(ctx).FirstOrCreate(&ing, domain.Ingredient{Name: name}).Error
	return &ing, err
}

func (r *recipeRepository) GetOrCreateSeasoning(ctx context.Context, name string) (*domain.Seasoning, error) {
	var sea domain.Seasoning
	err := r.db.WithContext(ctx).FirstOrCreate(&sea, domain.Seasoning{Name: name}).Error
	return &sea, err
}
