package repository

import (
	"context"
	"errors"

	"recipe-backend/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
func sharedRecipeIDs(db *gorm.DB, userID string) *gorm.DB {
	return db.Table("recipe_shares").
		Select("recipe_id").
		Where("user_id = ?", userID)
}

func (r *recipeRepository) FindAllForUser(ctx context.Context, userID string) ([]domain.Recipe, error) {
	var recipes []domain.Recipe
	db := r.db.WithContext(ctx)
	// そのユーザーの recipe_orders を LEFT JOIN し、position 昇順で並べる。
	// 並べ替え未設定(position が NULL)のレシピは末尾へ回し、id で安定させる。
	err := preloadAll(db).
		Joins("LEFT JOIN recipe_orders ON recipe_orders.recipe_id = recipes.id AND recipe_orders.user_id = ?", userID).
		Where("recipes.owner_id = ?", userID).
		Or("recipes.id IN (?)", sharedRecipeIDs(db, userID)).
		Order("recipe_orders.position ASC NULLS LAST").
		Order("recipes.id ASC").
		Find(&recipes).Error
	if err != nil {
		return nil, err
	}
	// 各レシピに、この userID にとってのアーカイブ状態を詰める。
	if err := r.fillArchived(ctx, userID, recipes); err != nil {
		return nil, err
	}
	return recipes, nil
}

// fillArchived は userID がアーカイブ済みのレシピを引いて、recipes の Archived を立てる。
func (r *recipeRepository) fillArchived(ctx context.Context, userID string, recipes []domain.Recipe) error {
	if len(recipes) == 0 {
		return nil
	}
	var archivedIDs []string
	if err := r.db.WithContext(ctx).
		Model(&domain.RecipeArchive{}).
		Where("user_id = ?", userID).
		Pluck("recipe_id", &archivedIDs).Error; err != nil {
		return err
	}
	archived := make(map[string]struct{}, len(archivedIDs))
	for _, id := range archivedIDs {
		archived[id] = struct{}{}
	}
	for i := range recipes {
		if _, ok := archived[recipes[i].ID]; ok {
			recipes[i].Archived = true
		}
	}
	return nil
}

func (r *recipeRepository) SetArchived(ctx context.Context, userID, recipeID string, archived bool) error {
	if !archived {
		return r.db.WithContext(ctx).
			Where("user_id = ? AND recipe_id = ?", userID, recipeID).
			Delete(&domain.RecipeArchive{}).Error
	}
	// 既に行があれば何もしない upsert。belongs-to(User/Recipe)は FK 定義用で、
	// 書き込みでは巻き込まないよう Omit する(recipe_orders と同じ扱い)。
	return r.db.WithContext(ctx).
		Omit("User", "Recipe").
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&domain.RecipeArchive{UserID: userID, RecipeID: recipeID}).Error
}

func (r *recipeRepository) IsArchived(ctx context.Context, userID, recipeID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.RecipeArchive{}).
		Where("user_id = ? AND recipe_id = ?", userID, recipeID).
		Count(&count).Error
	return count > 0, err
}

func (r *recipeRepository) Reorder(ctx context.Context, userID string, recipeIDs []string) error {
	if len(recipeIDs) == 0 {
		return nil
	}
	orders := make([]domain.RecipeOrder, 0, len(recipeIDs))
	for i, rid := range recipeIDs {
		orders = append(orders, domain.RecipeOrder{UserID: userID, RecipeID: rid, Position: i})
	}
	// (user_id, recipe_id) が既にあれば position を更新する upsert。
	// belongs-to(User/Recipe)は FK 定義用で、書き込みでは巻き込まないよう Omit する。
	return r.db.WithContext(ctx).
		Omit("User", "Recipe").
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "recipe_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"position"}),
		}).
		Create(&orders).Error
}

func (r *recipeRepository) FindByID(ctx context.Context, id string) (*domain.Recipe, error) {
	var recipe domain.Recipe
	err := preloadAll(r.db.WithContext(ctx)).Where("id = ?", id).First(&recipe).Error
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
		// アーカイブは per-user のため recipe_archives 側で扱い、ここでは触らない。
		// cooking_time は NULL を書ける必要があるため map で更新する。
		if err := tx.Model(&domain.Recipe{ID: recipe.ID}).Updates(map[string]any{
			"title":        recipe.Title,
			"cooking_time": recipe.CookingTime,
			"servings":     recipe.Servings,
			"procedure":    recipe.Procedure,
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
			recipe.Ingredients[i].ID = ""
			recipe.Ingredients[i].RecipeID = recipe.ID
		}
		if err := tx.Create(&recipe.Ingredients).Error; err != nil {
			return err
		}
	}
	if len(recipe.Seasonings) > 0 {
		for i := range recipe.Seasonings {
			recipe.Seasonings[i].ID = ""
			recipe.Seasonings[i].RecipeID = recipe.ID
		}
		if err := tx.Create(&recipe.Seasonings).Error; err != nil {
			return err
		}
	}
	if len(recipe.Labels) > 0 {
		for i := range recipe.Labels {
			recipe.Labels[i].ID = ""
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
