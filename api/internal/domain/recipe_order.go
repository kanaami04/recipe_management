package domain

// RecipeOrder はユーザーごとのレシピ表示順。(user, recipe) で 1 行を持ち、
// position が小さいほど前に並ぶ。並び順は各ユーザー個別で、所有・共有のどちらの
// レシピも各自が自由に並べ替えられ、他ユーザーの順序には影響しない。
//
// User / Recipe への belongs-to は ON DELETE CASCADE の FK を張るためだけに置く
// (書き込み時は関連の巻き込み作成を避けるため Omit する)。
type RecipeOrder struct {
	UserID   string `gorm:"type:uuid;primaryKey"`
	RecipeID string `gorm:"type:uuid;primaryKey"`
	Position int    `gorm:"not null"`

	User   User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Recipe Recipe `gorm:"foreignKey:RecipeID;constraint:OnDelete:CASCADE"`
}

func (RecipeOrder) TableName() string { return "recipe_orders" }
