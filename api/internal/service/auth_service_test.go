package service

import (
	"context"
	"testing"

	"recipe-backend/internal/domain"
	jwtpkg "recipe-backend/internal/pkg/jwt"
	"recipe-backend/internal/testutil/factory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newAuthService は mockUserRepo を組んだ AuthService を返すテストヘルパー。
// Login は email でユーザーを引くため、マップは email をキーにする。
func newAuthService(usersByEmail map[string]*domain.User) (AuthService, *jwtpkg.Manager) {
	ur := &mockUserRepo{byEmail: usersByEmail}
	jm := jwtpkg.NewManager("test-secret")
	return NewAuthService(ur, jm), jm
}

// loginAlice は有効ユーザー alice で正常ログインし、トークンと Manager を返す。
func loginAlice(t *testing.T) (access, refresh string, jm *jwtpkg.Manager) {
	t.Helper()
	user := factory.NewUser(factory.WithID("u1"), factory.WithEmail("alice@example.com"), factory.WithPlainPassword(t, "pw"))
	svc, jm := newAuthService(map[string]*domain.User{user.Email: user})
	access, refresh, err := svc.Login(context.Background(), "alice@example.com", "pw")
	require.NoError(t, err)
	return access, refresh, jm
}

// 正しい資格情報でログインした時、アクセストークンが返ること。
func TestAuthLogin_ReturnsAccessToken(t *testing.T) {
	// Arrange & Act: ログインはヘルパー内で実行される
	access, _, _ := loginAlice(t)

	// Assert
	assert.NotEmpty(t, access)
}

// 正しい資格情報でログインした時、リフレッシュトークンが返ること。
func TestAuthLogin_ReturnsRefreshToken(t *testing.T) {
	// Arrange & Act: ログインはヘルパー内で実行される
	_, refresh, _ := loginAlice(t)

	// Assert
	assert.NotEmpty(t, refresh)
}

// 正しい資格情報でログインした時、アクセストークンにユーザーIDが埋め込まれること。
func TestAuthLogin_AccessTokenEncodesUserID(t *testing.T) {
	// Arrange & Act: ログインはヘルパー内で実行される
	access, _, jm := loginAlice(t)

	// Assert
	uid, err := jm.Parse(access, jwtpkg.TypeAccess)
	require.NoError(t, err)
	assert.Equal(t, "u1", uid)
}

// パスワードが間違っている時、ErrInvalidCredentials が返ること。
func TestAuthLogin_WrongPassword(t *testing.T) {
	// Arrange
	user := factory.NewUser(factory.WithID("u1"), factory.WithEmail("alice@example.com"), factory.WithPlainPassword(t, "pw"))
	svc, _ := newAuthService(map[string]*domain.User{user.Email: user})

	// Act
	_, _, err := svc.Login(context.Background(), "alice@example.com", "wrong")

	// Assert
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

// ユーザーが存在しない時、ErrInvalidCredentials が返ること。
func TestAuthLogin_NoSuchUser(t *testing.T) {
	// Arrange
	svc, _ := newAuthService(map[string]*domain.User{})

	// Act
	_, _, err := svc.Login(context.Background(), "ghost@example.com", "pw")

	// Assert
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

// 無効化されたユーザーでログインした時、ErrInvalidCredentials が返ること。
func TestAuthLogin_InactiveUser(t *testing.T) {
	// Arrange
	user := factory.NewUser(factory.WithID("u1"), factory.WithEmail("alice@example.com"), factory.WithPlainPassword(t, "pw"), factory.WithInactive())
	svc, _ := newAuthService(map[string]*domain.User{user.Email: user})

	// Act
	_, _, err := svc.Login(context.Background(), "alice@example.com", "pw")

	// Assert
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

// 有効なリフレッシュトークンを渡した時、元の uid を保持したアクセストークンが返ること。
func TestAuthRefresh_Valid(t *testing.T) {
	// Arrange
	svc, jm := newAuthService(map[string]*domain.User{})
	refresh, _ := jm.GenerateRefresh("u5")

	// Act
	access, err := svc.Refresh(context.Background(), refresh)

	// Assert: 払い出した access が元の uid を保持していること
	require.NoError(t, err)
	uid, err := jm.Parse(access, jwtpkg.TypeAccess)
	require.NoError(t, err)
	assert.Equal(t, "u5", uid)
}

// リフレッシュとしてアクセストークンを渡した時、ErrInvalidCredentials が返ること。
func TestAuthRefresh_RejectsAccessToken(t *testing.T) {
	// Arrange
	svc, jm := newAuthService(map[string]*domain.User{})
	access, _ := jm.GenerateAccess("u5") // access を refresh として渡す → 失敗するはず

	// Act
	_, err := svc.Refresh(context.Background(), access)

	// Assert
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

// 不正な文字列をリフレッシュトークンに渡した時、ErrInvalidCredentials が返ること。
func TestAuthRefresh_Garbage(t *testing.T) {
	// Arrange
	svc, _ := newAuthService(map[string]*domain.User{})

	// Act
	_, err := svc.Refresh(context.Background(), "bad-token")

	// Assert
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

// registerAlice は alice を新規登録し、返り値ユーザーと保存先モックを返す。
func registerAlice(t *testing.T) (*domain.User, *mockUserRepo) {
	t.Helper()
	ur := &mockUserRepo{}
	svc := NewAuthService(ur, jwtpkg.NewManager("test-secret"))
	user, err := svc.Register(context.Background(), "alice", "alice@example.com", "password123")
	require.NoError(t, err)
	return user, ur
}

// 新規ユーザーを登録した時、ID が採番されること。
func TestAuthRegister_AssignsID(t *testing.T) {
	// Arrange & Act: 登録はヘルパー内で実行される
	user, _ := registerAlice(t)

	// Assert
	assert.NotZero(t, user.ID)
}

// 新規ユーザーを登録した時、パスワードがハッシュ化されて保存されること。
func TestAuthRegister_HashesPassword(t *testing.T) {
	// Arrange & Act: 登録はヘルパー内で実行される
	user, _ := registerAlice(t)

	// Assert
	assert.NotEqual(t, "password123", user.PasswordHash) // 平文ではなくハッシュ化されていること
}

// 新規ユーザーを登録した時、ユーザーが永続化されること。
func TestAuthRegister_PersistsUser(t *testing.T) {
	// Arrange & Act: 登録はヘルパー内で実行される
	_, ur := registerAlice(t)

	// Assert
	stored, _ := ur.FindByUsername(context.Background(), "alice")
	assert.NotNil(t, stored)
}

// username が既存と重複する時、ErrUserAlreadyExists が返ること。
func TestAuthRegister_DuplicateUsername(t *testing.T) {
	// Arrange
	existing := factory.NewUser(factory.WithID("u1"), factory.WithUsername("alice"))
	ur := &mockUserRepo{byName: map[string]*domain.User{existing.Username: existing}}
	svc := NewAuthService(ur, jwtpkg.NewManager("s"))

	// Act
	_, err := svc.Register(context.Background(), "alice", "new@example.com", "password123")

	// Assert
	assert.ErrorIs(t, err, ErrUserAlreadyExists)
}

// email が既存と重複する時、ErrUserAlreadyExists が返ること。
func TestAuthRegister_DuplicateEmail(t *testing.T) {
	// Arrange
	existing := factory.NewUser(factory.WithID("u1"), factory.WithEmail("taken@example.com"))
	ur := &mockUserRepo{byEmail: map[string]*domain.User{existing.Email: existing}}
	svc := NewAuthService(ur, jwtpkg.NewManager("s"))

	// Act
	_, err := svc.Register(context.Background(), "newuser", "taken@example.com", "password123")

	// Assert
	assert.ErrorIs(t, err, ErrUserAlreadyExists)
}
