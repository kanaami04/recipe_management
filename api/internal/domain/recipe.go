package domain

import "time"

// Recipe はレシピ本体。owner は所有者、SharedUsers は共有先ユーザー。
type Recipe struct {
	ID         uint   `gorm:"primaryKey"`
	Title      string `gorm:"size:50;not null"`
	CreateTime *int   // 調理時間（任意）
	CreateFor  int    `gorm:"default:1"` // 何人前
	Procedure  string `gorm:"type:text"`
	ArchiveFlg bool   `gorm:"default:false"`

	OwnerID uint            `gorm:"not null"`
	Owner   ApplicationUser `gorm:"foreignKey:OwnerID"`

	Labels      []RecipeLabel     `gorm:"many2many:recipes_label;joinForeignKey:RecipeID;joinReferences:RecipeLabelID"`
	SharedUsers []ApplicationUser `gorm:"many2many:recipes_shared_user;joinForeignKey:RecipeID;joinReferences:ApplicationUserID"`

	Cooking []Cooking `gorm:"foreignKey:RecipeID"`
	Season  []Season  `gorm:"foreignKey:RecipeID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Recipe) TableName() string { return "recipes" }
