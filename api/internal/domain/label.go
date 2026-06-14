package domain

// RecipeLabel はレシピ検索用のラベル。name は一意。
type RecipeLabel struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:50;uniqueIndex;not null"`
}

func (RecipeLabel) TableName() string { return "recipe_labels" }
