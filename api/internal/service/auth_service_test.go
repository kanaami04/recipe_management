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

// testVerifyURL / testResetURL はメールリンク組み立てに使うテスト用のベース URL。
const (
	testVerifyURL = "https://app.example.com/verify-email"
	testResetURL  = "https://app.example.com/reset-password/confirm"
)

// newAuthService は mockUserRepo を組んだ AuthService を返すテストヘルパー。
// Login は email でユーザーを引くため、マップは email をキーにする。
func newAuthService(usersByEmail map[string]*domain.User) (AuthService, *jwtpkg.Manager) {
	ur := &mockUserRepo{byEmail: usersByEmail}
	jm := jwtpkg.NewManager("test-secret")
	return NewAuthService(ur, jm, &mockMailer{}, testVerifyURL, testResetURL), jm
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
	svc := NewAuthService(ur, jwtpkg.NewManager("test-secret"), &mockMailer{}, testVerifyURL, testResetURL)
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
	svc := NewAuthService(ur, jwtpkg.NewManager("s"), &mockMailer{}, testVerifyURL, testResetURL)

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
	svc := NewAuthService(ur, jwtpkg.NewManager("s"), &mockMailer{}, testVerifyURL, testResetURL)

	// Act
	_, err := svc.Register(context.Background(), "newuser", "taken@example.com", "password123")

	// Assert
	assert.ErrorIs(t, err, ErrUserAlreadyExists)
}

// 新規登録した時、確認前(email_verified=false)で保存されること。
func TestAuthRegister_UnverifiedByDefault(t *testing.T) {
	// Arrange & Act
	user, _ := registerAlice(t)

	// Assert
	assert.False(t, user.EmailVerified)
}

// 新規登録した時、登録メール宛に確認メールが送られること。
func TestAuthRegister_SendsVerificationEmail(t *testing.T) {
	// Arrange
	ur := &mockUserRepo{}
	mailer := &mockMailer{}
	svc := NewAuthService(ur, jwtpkg.NewManager("test-secret"), mailer, testVerifyURL, testResetURL)

	// Act
	_, err := svc.Register(context.Background(), "alice", "alice@example.com", "password123")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", mailer.verifyTo)
	assert.Contains(t, mailer.verifyLink, testVerifyURL)
	assert.Contains(t, mailer.verifyLink, "token=")
}

// メール未確認のユーザーでログインした時、ErrEmailNotVerified が返ること。
func TestAuthLogin_Unverified(t *testing.T) {
	// Arrange
	user := factory.NewUser(factory.WithID("u1"), factory.WithEmail("alice@example.com"),
		factory.WithPlainPassword(t, "pw"), factory.WithEmailUnverified())
	svc, _ := newAuthService(map[string]*domain.User{user.Email: user})

	// Act
	_, _, err := svc.Login(context.Background(), "alice@example.com", "pw")

	// Assert
	assert.ErrorIs(t, err, ErrEmailNotVerified)
}

// 有効な確認トークン(発行時と現在のメールが一致)で検証した時、確認済みになること。
func TestAuthVerifyEmail_Valid(t *testing.T) {
	// Arrange
	user := factory.NewUser(factory.WithID("u9"), factory.WithEmail("alice@example.com"), factory.WithEmailUnverified())
	ur := &mockUserRepo{byID: map[string]*domain.User{user.ID: user}}
	jm := jwtpkg.NewManager("test-secret")
	svc := NewAuthService(ur, jm, &mockMailer{}, testVerifyURL, testResetURL)
	token, _ := jm.GenerateEmailVerify("u9", "alice@example.com")

	// Act
	err := svc.VerifyEmail(context.Background(), token)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, ur.emailVerifiedSet)
	assert.True(t, *ur.emailVerifiedSet)
}

// 発行後にメールを変えたユーザーが古いリンクを踏んだ時、現在のアドレスと一致せず弾かれること。
func TestAuthVerifyEmail_EmailChangedSinceIssued(t *testing.T) {
	// Arrange: トークンは旧アドレスで発行、DB 上の現在のアドレスは別物
	user := factory.NewUser(factory.WithID("u9"), factory.WithEmail("new@example.com"), factory.WithEmailUnverified())
	ur := &mockUserRepo{byID: map[string]*domain.User{user.ID: user}}
	jm := jwtpkg.NewManager("test-secret")
	svc := NewAuthService(ur, jm, &mockMailer{}, testVerifyURL, testResetURL)
	token, _ := jm.GenerateEmailVerify("u9", "old@example.com")

	// Act
	err := svc.VerifyEmail(context.Background(), token)

	// Assert
	assert.ErrorIs(t, err, ErrInvalidToken)
	assert.Nil(t, ur.emailVerifiedSet) // 確認済みにしない
}

// 確認トークンとして別種のトークンを渡した時、ErrInvalidToken が返ること。
func TestAuthVerifyEmail_WrongTokenType(t *testing.T) {
	// Arrange
	jm := jwtpkg.NewManager("test-secret")
	svc := NewAuthService(&mockUserRepo{}, jm, &mockMailer{}, testVerifyURL, testResetURL)
	access, _ := jm.GenerateAccess("u9") // access を確認トークンとして渡す

	// Act
	err := svc.VerifyEmail(context.Background(), access)

	// Assert
	assert.ErrorIs(t, err, ErrInvalidToken)
}

// 未確認ユーザーへ再送した時、確認メールが再度送られること。
func TestAuthResendVerification_Unverified(t *testing.T) {
	// Arrange
	user := factory.NewUser(factory.WithID("u1"), factory.WithEmail("alice@example.com"), factory.WithEmailUnverified())
	ur := &mockUserRepo{byEmail: map[string]*domain.User{user.Email: user}}
	mailer := &mockMailer{}
	svc := NewAuthService(ur, jwtpkg.NewManager("test-secret"), mailer, testVerifyURL, testResetURL)

	// Act
	err := svc.ResendVerification(context.Background(), "alice@example.com")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", mailer.verifyTo)
}

// 確認済み・不在ユーザーへ再送した時、メールを送らずエラーにもならないこと(列挙防止)。
func TestAuthResendVerification_NoOpForVerifiedOrMissing(t *testing.T) {
	// Arrange: 確認済み alice のみ
	verified := factory.NewUser(factory.WithID("u1"), factory.WithEmail("alice@example.com"))
	ur := &mockUserRepo{byEmail: map[string]*domain.User{verified.Email: verified}}
	mailer := &mockMailer{}
	svc := NewAuthService(ur, jwtpkg.NewManager("test-secret"), mailer, testVerifyURL, testResetURL)

	// Act & Assert: 確認済みでも不在でも、送信せず成功
	require.NoError(t, svc.ResendVerification(context.Background(), "alice@example.com"))
	require.NoError(t, svc.ResendVerification(context.Background(), "ghost@example.com"))
	assert.Empty(t, mailer.verifyTo)
}

// 存在するメールへリセット申請した時、リセットメールが送られること。
func TestAuthRequestPasswordReset_Existing(t *testing.T) {
	// Arrange
	user := factory.NewUser(factory.WithID("u1"), factory.WithEmail("alice@example.com"))
	ur := &mockUserRepo{byEmail: map[string]*domain.User{user.Email: user}}
	mailer := &mockMailer{}
	svc := NewAuthService(ur, jwtpkg.NewManager("test-secret"), mailer, testVerifyURL, testResetURL)

	// Act
	err := svc.RequestPasswordReset(context.Background(), "alice@example.com")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", mailer.resetTo)
	assert.Contains(t, mailer.resetLink, testResetURL)
}

// 不在のメールへリセット申請した時、メールを送らずエラーにもならないこと(列挙防止)。
func TestAuthRequestPasswordReset_MissingIsNoError(t *testing.T) {
	// Arrange
	mailer := &mockMailer{}
	svc := NewAuthService(&mockUserRepo{}, jwtpkg.NewManager("test-secret"), mailer, testVerifyURL, testResetURL)

	// Act
	err := svc.RequestPasswordReset(context.Background(), "ghost@example.com")

	// Assert
	require.NoError(t, err)
	assert.Empty(t, mailer.resetTo)
}

// 有効なリセットトークンで確定した時、対象ユーザーのパスワードが更新されること。
func TestAuthConfirmPasswordReset_Valid(t *testing.T) {
	// Arrange
	ur := &mockUserRepo{}
	jm := jwtpkg.NewManager("test-secret")
	svc := NewAuthService(ur, jm, &mockMailer{}, testVerifyURL, testResetURL)
	token, _ := jm.GeneratePasswordReset("u3")

	// Act
	err := svc.ConfirmPasswordReset(context.Background(), token, "newpassword")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "u3", ur.passwordChangedID)
	assert.NotEqual(t, "newpassword", ur.newPasswordHash) // ハッシュ化されていること
	require.NotNil(t, ur.emailVerifiedSet)
	assert.True(t, *ur.emailVerifiedSet) // リセット完了でメール到達性が証明されるため確認済みにする
}

// リセットトークンとして別種のトークンを渡した時、ErrInvalidToken が返ること。
func TestAuthConfirmPasswordReset_WrongTokenType(t *testing.T) {
	// Arrange
	jm := jwtpkg.NewManager("test-secret")
	svc := NewAuthService(&mockUserRepo{}, jm, &mockMailer{}, testVerifyURL, testResetURL)
	verify, _ := jm.GenerateEmailVerify("u3", "u3@example.com") // 確認トークンをリセットに流用

	// Act
	err := svc.ConfirmPasswordReset(context.Background(), verify, "newpassword")

	// Assert
	assert.ErrorIs(t, err, ErrInvalidToken)
}
