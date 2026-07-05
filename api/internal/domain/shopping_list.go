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

	// SharedUsers は「このリストを共有している相手」= グループのメンバー(owner 以外)。
	// 個別共有は廃止したため永続化せず(gorm:"-")、サービスがグループメンバーから詰める計算値。
	SharedUsers []User `gorm:"-"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ShoppingList) TableName() string { return "shopping_lists" }
