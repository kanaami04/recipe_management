package domain

import "time"

// ShareGroupMember はユーザーのシェアグループ所属。(group, user) で 1 行を持つ。
// 所有者自身もメンバー行を持つ。UserID に一意制約を張り、1 ユーザー 1 グループを担保する
// (将来 N:N に緩めるときは一意制約を外すだけでよい)。
//
// recipe_archives と同じく代理キーを持たない中間テーブル。User / ShareGroup への belongs-to は
// ON DELETE CASCADE の FK を張るためだけに置く(グループ削除・ユーザー削除で自動的に消える)。
type ShareGroupMember struct {
	GroupID string `gorm:"type:uuid;primaryKey"`
	UserID  string `gorm:"type:uuid;primaryKey;uniqueIndex"`

	Group    ShareGroup `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE"`
	User     User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	JoinedAt time.Time  `gorm:"autoCreateTime"`
}

func (ShareGroupMember) TableName() string { return "share_group_members" }
