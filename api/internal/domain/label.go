package domain

import "time"

// RecipeLabel はレシピに付いたラベル。レシピに従属し、(recipe, name) で一意。
// ラベル名はここに非正規化保持する(中間テーブルは作らない)。管理(作成・改名・削除)の
// 対象となるラベルの実体・一覧はユーザーごとの Label マスタが持つ。
type RecipeLabel struct {
	ID       string `gorm:"type:uuid;primaryKey"`
	RecipeID string `gorm:"type:uuid;not null;uniqueIndex:uniq_recipe_label_name"`
	Name     string `gorm:"size:50;not null;uniqueIndex:uniq_recipe_label_name"`
}

func (RecipeLabel) TableName() string { return "recipe_labels" }

// Label はユーザーが管理するラベル(マスタ)。作成・改名・削除の対象で、選択候補の source。
// レシピ側は RecipeLabel に name を非正規化保持し、改名・削除はそれへ name ベースで伝播する。
// 名前はユーザーごとに一意。
type Label struct {
	ID        string    `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"size:50;not null;uniqueIndex:uniq_label_owner_name"`
	OwnerID   string    `gorm:"type:uuid;not null;uniqueIndex:uniq_label_owner_name"`
	Owner     User      `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (Label) TableName() string { return "labels" }
