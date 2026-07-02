package domain

// RecipeLabel はレシピ検索用のラベル。レシピに従属し、(recipe, name) で一意。
// ラベル一覧はこのテーブルの name を DISTINCT で引く(マスタテーブルは持たない)。
type RecipeLabel struct {
	ID       uint   `gorm:"primaryKey"`
	RecipeID uint   `gorm:"not null;uniqueIndex:uniq_recipe_label_name"`
	Name     string `gorm:"size:50;not null;uniqueIndex:uniq_recipe_label_name"`
}

func (RecipeLabel) TableName() string { return "recipe_labels" }
