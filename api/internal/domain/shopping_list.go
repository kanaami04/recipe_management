package domain

import "time"

// ShoppingList は買い物リスト本体。1 ユーザー 1 リストを使い回す。Owner は所有者、
// SharedUsers は共有先ユーザー。Items はリストに従属する子テーブル(ON DELETE CASCADE)。
// 共有相手も同じ Items・同じ Checked 状態を共同編集する(世帯で 1 つのリストを共有する想定)。
type ShoppingList struct {
	ID string `gorm:"type:uuid;primaryKey"`
	// 1 ユーザーは所有リストを高々 1 つ(1 ユーザー 1 リストの使い回し)。同時作成の競合で
	// 二重に作られないよう owner_id に一意制約を張り、DB 側で不変式を担保する。
	OwnerID string `gorm:"type:uuid;not null;uniqueIndex"`
	Owner   User   `gorm:"foreignKey:OwnerID"`

	Items []ShoppingListItem `gorm:"constraint:OnDelete:CASCADE"`

	// リストとユーザーの多対多(共有)。中間テーブル shopping_list_shares は
	// many2many 定義から自動生成される(recipe_shares と同型)。
	SharedUsers []User `gorm:"many2many:shopping_list_shares;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ShoppingList) TableName() string { return "shopping_lists" }
