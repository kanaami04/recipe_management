package domain

import "time"

// ApplicationUser はユーザーテーブル。
type ApplicationUser struct {
	ID          uint      `gorm:"primaryKey"`
	Username    string    `gorm:"size:50;uniqueIndex;not null"`
	Email       string    `gorm:"size:50;uniqueIndex;not null"`
	Password    string    `gorm:"not null"`
	IsStaff     bool      `gorm:"default:false"`
	IsSuperuser bool      `gorm:"default:false"`
	IsActive    bool      `gorm:"default:true"`
	DateJoined  time.Time `gorm:"autoCreateTime"`
}

func (ApplicationUser) TableName() string { return "application_user" }
