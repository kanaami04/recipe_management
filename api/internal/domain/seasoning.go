package domain

// Seasoning は調味料。
type Seasoning struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:50;not null"`
}

func (Seasoning) TableName() string { return "seasoning" }

// Season はレシピと調味料の中間テーブル。(recipe, seasoning) で一意。
type Season struct {
	ID          uint      `gorm:"primaryKey"`
	RecipeID    uint      `gorm:"not null;uniqueIndex:uniq_recipe_seasoning"`
	SeasoningID uint      `gorm:"not null;uniqueIndex:uniq_recipe_seasoning"`
	Seasoning   Seasoning `gorm:"foreignKey:SeasoningID"`
	Quantity    int       `gorm:"not null"`
	Unit        string    `gorm:"size:10;not null"`
}

func (Season) TableName() string { return "season" }
