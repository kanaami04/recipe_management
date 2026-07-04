package domain

// RecipeArchive はユーザーごとのレシピのアーカイブ状態。(user, recipe) で 1 行を持ち、
// 行が存在すれば「そのユーザーにとってアーカイブ済み」を表す。アーカイブは各ユーザー個別で、
// 所有・共有のどちらのレシピも各自が自由にアーカイブでき、他ユーザーの状態には影響しない。
//
// recipe_orders と同じく、entity 間の関連ではなく「ユーザー固有の状態」を持つテーブル。
// User / Recipe への belongs-to は ON DELETE CASCADE の FK を張るためだけに置く
// (書き込み時は関連の巻き込み作成を避けるため Omit する)。
type RecipeArchive struct {
	UserID   string `gorm:"type:uuid;primaryKey"`
	RecipeID string `gorm:"type:uuid;primaryKey"`

	User   User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Recipe Recipe `gorm:"foreignKey:RecipeID;constraint:OnDelete:CASCADE"`
}

func (RecipeArchive) TableName() string { return "recipe_archives" }
