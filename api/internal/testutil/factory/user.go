// Package factory はテストデータ生成用のファクトリを提供する。
// functional options パターンで、デフォルト値を持ちつつ必要な属性だけ上書きできる。
package factory

import (
	"testing"

	"recipe-backend/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

// UserOption は NewUser の生成オプション。
type UserOption func(*domain.ApplicationUser)

// NewUser はテスト用の ApplicationUser を生成する。
// デフォルトは「有効な一般ユーザー」。必要な属性だけオプションで上書きする。
func NewUser(opts ...UserOption) *domain.ApplicationUser {
	u := &domain.ApplicationUser{
		Username: "testuser",
		Email:    "testuser@example.com",
		Password: "x",
		IsActive: true,
	}
	for _, opt := range opts {
		opt(u)
	}
	return u
}

// WithID は ID を指定する。
func WithID(id uint) UserOption {
	return func(u *domain.ApplicationUser) { u.ID = id }
}

// WithUsername は username を指定する。
func WithUsername(name string) UserOption {
	return func(u *domain.ApplicationUser) { u.Username = name }
}

// WithEmail は email を指定する。
func WithEmail(email string) UserOption {
	return func(u *domain.ApplicationUser) { u.Email = email }
}

// WithInactive は無効ユーザー（IsActive=false）にする。
func WithInactive() UserOption {
	return func(u *domain.ApplicationUser) { u.IsActive = false }
}

// WithPlainPassword は平文パスワードを bcrypt でハッシュ化してセットする。
// ログイン系テストで「正しいパスワード」を用意するために使う。
func WithPlainPassword(t *testing.T, pw string) UserOption {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("パスワードのハッシュ化に失敗: %v", err)
	}
	return func(u *domain.ApplicationUser) { u.Password = string(hash) }
}
