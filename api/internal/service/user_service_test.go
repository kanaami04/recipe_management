package service

import (
	"context"
	"testing"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/testutil/factory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 該当ユーザーがいる時、GetByID でそのユーザーが構造体ごと返ること。
func TestUserGetByID_Found(t *testing.T) {
	// Arrange
	user := factory.NewUser(factory.WithID("u7"), factory.WithUsername("alice"))
	ur := &mockUserRepo{byID: map[string]*domain.User{"u7": user}}
	svc := NewUserService(ur)

	// Act
	got, err := svc.GetByID(context.Background(), "u7")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, user, got)
}

// 該当ユーザーがいない時、GetByID で nil が返ること。
func TestUserGetByID_NotFound(t *testing.T) {
	// Arrange
	ur := &mockUserRepo{byID: map[string]*domain.User{}}
	svc := NewUserService(ur)

	// Act
	got, err := svc.GetByID(context.Background(), "u999")

	// Assert
	require.NoError(t, err)
	assert.Nil(t, got)
}

// ユーザーが登録されている時、List で全件が返ること。
func TestUserList_ReturnsAll(t *testing.T) {
	// Arrange
	ur := &mockUserRepo{all: []domain.User{
		*factory.NewUser(factory.WithID("u1"), factory.WithUsername("alice")),
		*factory.NewUser(factory.WithID("u2"), factory.WithUsername("bob")),
	}}
	svc := NewUserService(ur)

	// Act
	users, err := svc.List(context.Background())

	// Assert
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

// ユーザーが1人もいない時、List で空が返ること。
func TestUserList_Empty(t *testing.T) {
	// Arrange
	svc := NewUserService(&mockUserRepo{})

	// Act
	users, err := svc.List(context.Background())

	// Assert
	require.NoError(t, err)
	assert.Empty(t, users)
}

// arrangeSelfUser は自分(u1)を byID/byName/byEmail に登録した mock を返す。
func arrangeSelfUser(opts ...factory.UserOption) *mockUserRepo {
	base := []factory.UserOption{
		factory.WithID("u1"), factory.WithUsername("alice"), factory.WithEmail("alice@example.com"),
	}
	self := factory.NewUser(append(base, opts...)...)
	return &mockUserRepo{
		byID:    map[string]*domain.User{self.ID: self},
		byName:  map[string]*domain.User{self.Username: self},
		byEmail: map[string]*domain.User{self.Email: self},
	}
}

// プロフィールを更新した時、新しい username がリポジトリへ渡ること。
func TestUserUpdateProfile_SavesUsername(t *testing.T) {
	// Arrange
	ur := arrangeSelfUser()
	svc := NewUserService(ur)

	// Act
	_, err := svc.UpdateProfile(context.Background(), "u1", "alice2")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, ur.updated)
	assert.Equal(t, "alice2", ur.updated.Username)
}

// 別ユーザーが使っている username に変更した時、ErrUserAlreadyExists が返ること。
func TestUserUpdateProfile_DuplicateUsername(t *testing.T) {
	// Arrange: bob(u2)が既にいる
	ur := arrangeSelfUser()
	ur.byName["bob"] = factory.NewUser(factory.WithID("u2"), factory.WithUsername("bob"))
	svc := NewUserService(ur)

	// Act
	_, err := svc.UpdateProfile(context.Background(), "u1", "bob")

	// Assert
	assert.ErrorIs(t, err, ErrUserAlreadyExists)
}

// 値を変えずに更新した時、自分自身との重複判定で弾かれず成功すること。
func TestUserUpdateProfile_SameValue(t *testing.T) {
	// Arrange
	ur := arrangeSelfUser()
	svc := NewUserService(ur)

	// Act
	_, err := svc.UpdateProfile(context.Background(), "u1", "alice")

	// Assert
	require.NoError(t, err)
}

// 正しいパスワードでメールを変更した時、新しい email がリポジトリへ渡ること。
func TestUserChangeEmail_Saves(t *testing.T) {
	// Arrange
	ur := arrangeSelfUser(factory.WithPlainPassword(t, "pass1234"))
	svc := NewUserService(ur)

	// Act
	_, err := svc.ChangeEmail(context.Background(), "u1", "alice2@example.com", "pass1234")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, ur.updated)
	assert.Equal(t, "alice2@example.com", ur.updated.Email)
}

// パスワードが違う時、メール変更が ErrIncorrectPassword で弾かれること。
func TestUserChangeEmail_WrongPassword(t *testing.T) {
	// Arrange
	ur := arrangeSelfUser(factory.WithPlainPassword(t, "pass1234"))
	svc := NewUserService(ur)

	// Act
	_, err := svc.ChangeEmail(context.Background(), "u1", "alice2@example.com", "WRONG")

	// Assert
	assert.ErrorIs(t, err, ErrIncorrectPassword)
}

// 別ユーザーが使っている email に変更した時、ErrUserAlreadyExists が返ること。
func TestUserChangeEmail_Duplicate(t *testing.T) {
	// Arrange
	ur := arrangeSelfUser(factory.WithPlainPassword(t, "pass1234"))
	ur.byEmail["bob@example.com"] = factory.NewUser(factory.WithID("u2"), factory.WithEmail("bob@example.com"))
	svc := NewUserService(ur)

	// Act
	_, err := svc.ChangeEmail(context.Background(), "u1", "bob@example.com", "pass1234")

	// Assert
	assert.ErrorIs(t, err, ErrUserAlreadyExists)
}

// 現在のパスワードが正しい時、対象ユーザーのパスワードが更新されること。
func TestUserChangePassword_UpdatesForUser(t *testing.T) {
	// Arrange
	ur := arrangeSelfUser(factory.WithPlainPassword(t, "oldpass12"))
	svc := NewUserService(ur)

	// Act
	err := svc.ChangePassword(context.Background(), "u1", "oldpass12", "newpass34")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "u1", ur.passwordChangedID)
}

// 現在のパスワードが正しい時、新しいパスワードがハッシュ化されて渡ること。
func TestUserChangePassword_HashesNewPassword(t *testing.T) {
	// Arrange
	ur := arrangeSelfUser(factory.WithPlainPassword(t, "oldpass12"))
	svc := NewUserService(ur)

	// Act
	err := svc.ChangePassword(context.Background(), "u1", "oldpass12", "newpass34")

	// Assert: 平文ではなくハッシュ化された値が保存される
	require.NoError(t, err)
	assert.NotEqual(t, "newpass34", ur.newPasswordHash)
}

// 現在のパスワードが違う時、ErrIncorrectPassword が返ること。
func TestUserChangePassword_WrongCurrent(t *testing.T) {
	// Arrange
	ur := arrangeSelfUser(factory.WithPlainPassword(t, "oldpass12"))
	svc := NewUserService(ur)

	// Act
	err := svc.ChangePassword(context.Background(), "u1", "wrongpass", "newpass34")

	// Assert
	assert.ErrorIs(t, err, ErrIncorrectPassword)
}

// アカウントを削除した時、その userID がリポジトリへ渡ること。
func TestUserDeleteAccount_Deletes(t *testing.T) {
	// Arrange
	ur := arrangeSelfUser()
	svc := NewUserService(ur)

	// Act
	err := svc.DeleteAccount(context.Background(), "u1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "u1", ur.deletedUserID)
}
