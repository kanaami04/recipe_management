package domain

import "time"

// ShareGroup はシェアグループ(世帯)。グループのメンバーは、各自が所有するレシピ・買い物
// リストなど共有対象を自動的に全員で共有し合う。1 ユーザーは高々 1 グループに所属する
// (share_group_members.user_id の一意制約で担保)。
//
// Owner はグループを作成した管理者で、メンバー削除・解散・招待コード再発行ができる。
// Members は share_group_members から詰める計算値(所有者自身もメンバー行を持つ)。
type ShareGroup struct {
	ID      string `gorm:"type:uuid;primaryKey"`
	Name    string `gorm:"size:50;not null"`
	OwnerID string `gorm:"type:uuid;not null;index"`
	Owner   User   `gorm:"foreignKey:OwnerID"`

	// InviteCode は参加用の招待コード。再発行で古いコードを失効させる(コードを差し替える)。
	InviteCode          string    `gorm:"size:16;uniqueIndex;not null"`
	InviteCodeExpiresAt time.Time `gorm:"not null"`

	// Members は share_group_members から詰める計算値(永続化しない)。
	Members []User `gorm:"-"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ShareGroup) TableName() string { return "share_groups" }
