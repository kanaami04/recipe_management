package domain

// RecipeIngredient はレシピの材料(食材)。レシピに従属し、(recipe, name) で一意。
// 食材マスタは持たず、名前をそのまま保持する(非正規化)。
type RecipeIngredient struct {
	ID       string `gorm:"type:uuid;primaryKey"`
	RecipeID string `gorm:"type:uuid;not null;uniqueIndex:uniq_recipe_ingredient_name"`
	Name     string `gorm:"size:50;not null;uniqueIndex:uniq_recipe_ingredient_name"`
	Quantity int    `gorm:"not null"`
	Unit     string `gorm:"size:10;not null"`
}

func (RecipeIngredient) TableName() string { return "recipe_ingredients" }
