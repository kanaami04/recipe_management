package domain

import "time"

// User はユーザー。
// テーブル名は user が PostgreSQL の予約語のため複数形の users とする。
type User struct {
	ID           string `gorm:"type:uuid;primaryKey"`
	Username     string `gorm:"size:50;uniqueIndex;not null"`
	Email        string `gorm:"size:50;uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	IsActive     bool   `gorm:"not null;default:true"`
	// EmailVerified はメールアドレスの実在確認が済んでいるか。false の間はログインできない
	//(auth_service の Login で弾く)。列追加時、既存ユーザーは移行で true 埋めする(後方互換)。
	EmailVerified bool `gorm:"not null;default:false"`
	// AvatarKey はプロフィール画像のオブジェクトストレージ(S3 互換)上のキー。未設定なら nil。
	AvatarKey *string   `gorm:"size:255"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (User) TableName() string { return "users" }
