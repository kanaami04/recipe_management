package service

import (
	"context"
	"strings"
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
	svc := NewUserService(ur, &mockAvatarStorage{})

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
	svc := NewUserService(ur, &mockAvatarStorage{})

	// Act
	got, err := svc.GetByID(context.Background(), "u999")

	// Assert
	require.NoError(t, err)
	assert.Nil(t, got)
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
	svc := NewUserService(ur, &mockAvatarStorage{})

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
	svc := NewUserService(ur, &mockAvatarStorage{})

	// Act
	_, err := svc.UpdateProfile(context.Background(), "u1", "bob")

	// Assert
	assert.ErrorIs(t, err, ErrUserAlreadyExists)
}

// 値を変えずに更新した時、自分自身との重複判定で弾かれず成功すること。
func TestUserUpdateProfile_SameValue(t *testing.T) {
	// Arrange
	ur := arrangeSelfUser()
	svc := NewUserService(ur, &mockAvatarStorage{})

	// Act
	_, err := svc.UpdateProfile(context.Background(), "u1", "alice")

	// Assert
	require.NoError(t, err)
}

// 正しいパスワードでメールを変更した時、新しい email がリポジトリへ渡ること。
func TestUserChangeEmail_Saves(t *testing.T) {
	// Arrange
	ur := arrangeSelfUser(factory.WithPlainPassword(t, "pass1234"))
	svc := NewUserService(ur, &mockAvatarStorage{})

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
	svc := NewUserService(ur, &mockAvatarStorage{})

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
	svc := NewUserService(ur, &mockAvatarStorage{})

	// Act
	_, err := svc.ChangeEmail(context.Background(), "u1", "bob@example.com", "pass1234")

	// Assert
	assert.ErrorIs(t, err, ErrUserAlreadyExists)
}

// 現在のパスワードが正しい時、対象ユーザーのパスワードが更新されること。
func TestUserChangePassword_UpdatesForUser(t *testing.T) {
	// Arrange
	ur := arrangeSelfUser(factory.WithPlainPassword(t, "oldpass12"))
	svc := NewUserService(ur, &mockAvatarStorage{})

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
	svc := NewUserService(ur, &mockAvatarStorage{})

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
	svc := NewUserService(ur, &mockAvatarStorage{})

	// Act
	err := svc.ChangePassword(context.Background(), "u1", "wrongpass", "newpass34")

	// Assert
	assert.ErrorIs(t, err, ErrIncorrectPassword)
}

// アカウントを削除した時、その userID がリポジトリへ渡ること。
func TestUserDeleteAccount_Deletes(t *testing.T) {
	// Arrange
	ur := arrangeSelfUser()
	svc := NewUserService(ur, &mockAvatarStorage{})

	// Act
	err := svc.DeleteAccount(context.Background(), "u1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "u1", ur.deletedUserID)
}

// アバターのアップロード URL を発行した時、自分の接頭辞を持つ key がストレージに渡ること。
func TestUserCreateAvatarUploadURL_KeyIsOwnedPrefix(t *testing.T) {
	// Arrange
	ur := arrangeSelfUser()
	st := &mockAvatarStorage{}
	svc := NewUserService(ur, st)

	// Act
	_, key, err := svc.CreateAvatarUploadURL(context.Background(), "u1", "image/png")

	// Assert: key は avatars/{userID}/ 配下(ConfirmAvatar の所有チェックと一致)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(key, "avatars/u1/"))
}

// 自分宛ての key を確定した時、その key が avatar_key に保存されること。
func TestUserConfirmAvatar_SavesKey(t *testing.T) {
	// Arrange
	ur := arrangeSelfUser()
	svc := NewUserService(ur, &mockAvatarStorage{})
	key := "avatars/u1/abc"

	// Act
	_, err := svc.ConfirmAvatar(context.Background(), "u1", key)

	// Assert
	require.NoError(t, err)
	require.Len(t, ur.avatarKeyUpdates, 1)
	require.NotNil(t, ur.avatarKeyUpdates[0])
	assert.Equal(t, key, *ur.avatarKeyUpdates[0])
}

// 他人宛ての key を確定しようとした時、ErrForbidden が返り保存されないこと。
func TestUserConfirmAvatar_ForbiddenForOthersKey(t *testing.T) {
	// Arrange
	ur := arrangeSelfUser()
	svc := NewUserService(ur, &mockAvatarStorage{})

	// Act: 別ユーザー(u2)配下の key
	_, err := svc.ConfirmAvatar(context.Background(), "u1", "avatars/u2/abc")

	// Assert
	assert.ErrorIs(t, err, ErrForbidden)
}

// 既存アバターがある状態で確定した時、古い画像がストレージから削除されること。
func TestUserConfirmAvatar_DeletesOldObject(t *testing.T) {
	// Arrange: 既にアバターを持つ
	old := "avatars/u1/old"
	ur := arrangeSelfUser()
	ur.byID["u1"].AvatarKey = &old
	st := &mockAvatarStorage{}
	svc := NewUserService(ur, st)

	// Act
	_, err := svc.ConfirmAvatar(context.Background(), "u1", "avatars/u1/new")

	// Assert
	require.NoError(t, err)
	assert.Contains(t, st.deletedKeys, old)
}

// 同じ key を二重に確定した時、現在の画像オブジェクトを消さないこと。
func TestUserConfirmAvatar_SameKeyKeepsObject(t *testing.T) {
	// Arrange: 現在のアバターと同じ key を確定する
	key := "avatars/u1/same"
	ur := arrangeSelfUser()
	ur.byID["u1"].AvatarKey = &key
	st := &mockAvatarStorage{}
	svc := NewUserService(ur, st)

	// Act
	_, err := svc.ConfirmAvatar(context.Background(), "u1", key)

	// Assert: 現行 key は削除されない
	require.NoError(t, err)
	assert.NotContains(t, st.deletedKeys, key)
}

// arrangeUserWithAvatar は avatar を持つ u1 の mock repo とストレージ mock を返す。
func arrangeUserWithAvatar(key string) (*mockUserRepo, *mockAvatarStorage) {
	ur := arrangeSelfUser()
	ur.byID["u1"].AvatarKey = &key
	return ur, &mockAvatarStorage{}
}

// アバターを削除した時、ストレージからオブジェクトが消えること。
func TestUserDeleteAvatar_DeletesObject(t *testing.T) {
	// Arrange
	key := "avatars/u1/abc"
	ur, st := arrangeUserWithAvatar(key)
	svc := NewUserService(ur, st)

	// Act
	_, err := svc.DeleteAvatar(context.Background(), "u1")

	// Assert
	require.NoError(t, err)
	assert.Contains(t, st.deletedKeys, key)
}

// アバターを削除した時、avatar_key が nil に戻ること。
func TestUserDeleteAvatar_ClearsKey(t *testing.T) {
	// Arrange
	ur, st := arrangeUserWithAvatar("avatars/u1/abc")
	svc := NewUserService(ur, st)

	// Act
	_, err := svc.DeleteAvatar(context.Background(), "u1")

	// Assert
	require.NoError(t, err)
	require.Len(t, ur.avatarKeyUpdates, 1)
	assert.Nil(t, ur.avatarKeyUpdates[0])
}
