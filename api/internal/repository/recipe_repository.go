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
// SharedUsers はシェアグループのメンバーから service が詰める計算値のため preload しない。
func preloadAll(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Owner").
		Preload("Labels").
		Preload("Ingredients").
		Preload("Seasonings")
}

// sameGroupOwnerIDs は userID と同じシェアグループに属する全ユーザー ID のサブクエリを返す
// (自分を含む)。どのグループにも属さなければ空。レシピの可視範囲の算出に使う。
func sameGroupOwnerIDs(db *gorm.DB, userID string) *gorm.DB {
	return db.Table("share_group_members AS m2").
		Select("m2.user_id").
		Joins("JOIN share_group_members AS m1 ON m1.group_id = m2.group_id").
		Where("m1.user_id = ?", userID)
}

func (r *recipeRepository) FindAllForUser(ctx context.Context, userID string) ([]domain.Recipe, error) {
	var recipes []domain.Recipe
	db := r.db.WithContext(ctx)
	// 可視範囲は「自分が所有」または「同じシェアグループのメンバーが所有」するレシピ。
	// そのユーザーの recipe_orders を LEFT JOIN し、position 昇順で並べる。
	// 並べ替え未設定(position が NULL)のレシピは末尾へ回し、id で安定させる。
	err := preloadAll(db).
		Joins("LEFT JOIN recipe_orders ON recipe_orders.recipe_id = recipes.id AND recipe_orders.user_id = ?", userID).
		Where("recipes.owner_id = ?", userID).
		Or("recipes.owner_id IN (?)", sameGroupOwnerIDs(db, userID)).
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

// orphanRecipeIDs は userID が今は見られないレシピ(自分の所有でも、同じシェアグループの
// メンバー所有でもないレシピ)の ID を引くサブクエリを返す。掃除対象の絞り込みに使う。
func orphanRecipeIDs(db *gorm.DB, userID string) *gorm.DB {
	return db.Table("recipes").
		Select("recipes.id").
		Where("recipes.owner_id <> ?", userID).
		Where("recipes.owner_id NOT IN (?)", sameGroupOwnerIDs(db, userID))
}

func (r *recipeRepository) PruneRecipeState(ctx context.Context, userID string) error {
	// 見えなくなったレシピに対する per-user 状態(アーカイブ・並び順)を両方消す。
	// 2 テーブルへの削除をまとめて成否させるためトランザクションで包む。
	// サブクエリは実行しないため 2 文で作り直す(可視範囲は呼び出し時点のメンバー行に依る)。
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Where("user_id = ? AND recipe_id IN (?)", userID, orphanRecipeIDs(tx, userID)).
			Delete(&domain.RecipeArchive{}).Error; err != nil {
			return err
		}
		return tx.
			Where("user_id = ? AND recipe_id IN (?)", userID, orphanRecipeIDs(tx, userID)).
			Delete(&domain.RecipeOrder{}).Error
	})
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
	// 子テーブル(Labels/Ingredients/Seasonings)は本体と一緒に insert される。
	// 共有はシェアグループで管理するため、レシピ側は owner のみ持つ。
	return r.db.WithContext(ctx).Omit("Owner").Create(recipe).Error
}

func (r *recipeRepository) Update(ctx context.Context, recipe *domain.Recipe) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// スカラー項目のみ更新（owner / created_at は変更しない）。
		// アーカイブは per-user のため recipe_archives 側で扱い、ここでは触らない。
		// cooking_time は NULL を書ける必要があるため map で更新する。
		if err := tx.Model(&domain.Recipe{ID: recipe.ID}).Updates(map[string]any{
			"title":         recipe.Title,
			"cooking_time":  recipe.CookingTime,
			"servings":      recipe.Servings,
			"procedure":     recipe.Procedure,
			"source_url":    recipe.SourceURL,
			"thumbnail_url": recipe.ThumbnailURL,
		}).Error; err != nil {
			return err
		}

		// 子テーブルは全置換（差分更新はしない）。共有はグループ管理のため触らない。
		return replaceChildren(tx, recipe)
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
	// 子テーブル(材料・調味料・ラベル)は FK の ON DELETE CASCADE で削除される。
	return r.db.WithContext(ctx).Delete(recipe).Error
}
