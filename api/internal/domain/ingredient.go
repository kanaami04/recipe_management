package domain

// Ingredient は食材。
type Ingredient struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:50;not null"`
}

func (Ingredient) TableName() string { return "ingredients" }

// Cooking はレシピと食材の中間テーブル。(recipe, ingredient) で一意。
type Cooking struct {
	ID           uint       `gorm:"primaryKey"`
	RecipeID     uint       `gorm:"not null;uniqueIndex:uniq_recipe_ingredient"`
	IngredientID uint       `gorm:"not null;uniqueIndex:uniq_recipe_ingredient"`
	Ingredient   Ingredient `gorm:"foreignKey:IngredientID"`
	Quantity     int        `gorm:"not null"`
	Unit         string     `gorm:"size:10;not null"`
}

func (Cooking) TableName() string { return "cooking" }
