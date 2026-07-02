package domain

import "time"

// User はユーザー。
// テーブル名は user が PostgreSQL の予約語のため複数形の users とする。
type User struct {
	ID           uint      `gorm:"primaryKey"`
	Username     string    `gorm:"size:50;uniqueIndex;not null"`
	Email        string    `gorm:"size:50;uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	IsActive     bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func (User) TableName() string { return "users" }
