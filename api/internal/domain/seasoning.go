package domain

// RecipeSeasoning はレシピの調味料。レシピに従属し、(recipe, name) で一意。
// 調味料マスタは持たず、名前をそのまま保持する(非正規化)。
type RecipeSeasoning struct {
	ID       uint   `gorm:"primaryKey"`
	RecipeID uint   `gorm:"not null;uniqueIndex:uniq_recipe_seasoning_name"`
	Name     string `gorm:"size:50;not null;uniqueIndex:uniq_recipe_seasoning_name"`
	Quantity int    `gorm:"not null"`
	Unit     string `gorm:"size:10;not null"`
}

func (RecipeSeasoning) TableName() string { return "recipe_seasonings" }
