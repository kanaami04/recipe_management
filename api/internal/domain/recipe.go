package domain

import "time"

// Recipe はレシピ本体。Owner は所有者、SharedUsers は共有先ユーザー。
// 材料・調味料・ラベルはレシピに従属する子テーブル(ON DELETE CASCADE)。
type Recipe struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"size:50;not null"`
	CookingTime *int   // 調理時間(分)。任意
	Servings    int    `gorm:"not null;default:1"` // 何人前
	Procedure   string `gorm:"type:text"`
	Archived    bool   `gorm:"not null;default:false"`

	OwnerID uint `gorm:"not null;index"`
	Owner   User `gorm:"foreignKey:OwnerID"`

	Ingredients []RecipeIngredient `gorm:"constraint:OnDelete:CASCADE"`
	Seasonings  []RecipeSeasoning  `gorm:"constraint:OnDelete:CASCADE"`
	Labels      []RecipeLabel      `gorm:"constraint:OnDelete:CASCADE"`

	// レシピとユーザーの多対多(共有)。ここだけは中間テーブル(recipe_shares)が必要。
	SharedUsers []User `gorm:"many2many:recipe_shares;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Recipe) TableName() string { return "recipes" }
