package domain

// ShoppingListItem は買い物リストの項目。リストに従属する(ON DELETE CASCADE)。
// Checked は共有相手全員で共有する 1 つの状態。Position は手動並び替えの表示順
// (小さいほど上)で、これも共有相手全員で共有する。
// 一覧は Checked 昇順 → Position 昇順 → ID 昇順で並べ、チェック済みは末尾へ回す。
//
// Quantity / Unit はレシピから追加したときの分量(例: 200 / "ml")。手動追加した
// 項目は分量を持たず Quantity は nil・Unit は空文字になる。
type ShoppingListItem struct {
	ID             string   `gorm:"type:uuid;primaryKey"`
	ShoppingListID string   `gorm:"type:uuid;not null;index"`
	Name           string   `gorm:"size:50;not null"`
	Quantity       *float64 `gorm:"type:numeric"`
	Unit           string   `gorm:"size:10;not null;default:''"`
	Checked        bool     `gorm:"not null;default:false"`
	Position       int      `gorm:"not null;default:0"`
}

func (ShoppingListItem) TableName() string { return "shopping_list_items" }
