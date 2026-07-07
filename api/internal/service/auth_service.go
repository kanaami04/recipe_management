package service

import (
	"context"
	"fmt"
	"log/slog"

	"recipe-backend/internal/domain"
	jwtpkg "recipe-backend/internal/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (access, refresh string, err error)
	Refresh(ctx context.Context, refreshToken string) (access string, err error)
	Register(ctx context.Context, username, email, password string) (*domain.User, error)
	// VerifyEmail はメール確認トークンを検証し、有効なら email_verified を true にする。
	VerifyEmail(ctx context.Context, token string) error
	// ResendVerification は email のユーザーが未確認なら確認メールを再送する。
	// メール列挙を防ぐため、ユーザー不在・確認済みでもエラーにせず何もしないで返す。
	ResendVerification(ctx context.Context, email string) error
	// RequestPasswordReset は email のユーザーがいればリセットメールを送る。
	// メール列挙を防ぐため、ユーザー不在でもエラーにせず返す。
	RequestPasswordReset(ctx context.Context, email string) error
	// ConfirmPasswordReset はリセットトークンを検証し、有効なら password_hash を更新する。
	ConfirmPasswordReset(ctx context.Context, token, newPassword string) error
}

type authService struct {
	users  domain.UserRepository
	jwt    *jwtpkg.Manager
	mailer domain.Mailer
	// emailVerifyURL / passwordResetURL はメール内リンクのベース URL(フロントのページ)。
	emailVerifyURL   string
	passwordResetURL string
}

func NewAuthService(
	users domain.UserRepository,
	jwt *jwtpkg.Manager,
	mailer domain.Mailer,
	emailVerifyURL, passwordResetURL string,
) AuthService {
	return &authService{
		users:            users,
		jwt:              jwt,
		mailer:           mailer,
		emailVerifyURL:   emailVerifyURL,
		passwordResetURL: passwordResetURL,
	}
}

func (s *authService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}
	if user == nil || !user.IsActive {
		return "", "", ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", "", ErrInvalidCredentials
	}
	// パスワード一致後に未確認を弾く(未確認かどうかで認証情報の当たり外れが漏れないよう
	// 資格情報検証の後に判定する)。
	if !user.EmailVerified {
		return "", "", ErrEmailNotVerified
	}

	access, err := s.jwt.GenerateAccess(user.ID)
	if err != nil {
		return "", "", err
	}
	refresh, err := s.jwt.GenerateRefresh(user.ID)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (string, error) {
	userID, err := s.jwt.Parse(refreshToken, jwtpkg.TypeRefresh)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	return s.jwt.GenerateAccess(userID)
}

// Register は新規ユーザーを作成し、確認メールを送る。username/email の重複は ErrUserAlreadyExists。
// 作成したユーザーは email_verified=false のままで、確認するまでログインできない。
func (s *authService) Register(ctx context.Context, username, email, password string) (*domain.User, error) {
	if existing, err := s.users.FindByUsername(ctx, username); err != nil {
		return nil, err
	} else if existing != nil {
		return nil, fmt.Errorf("%w: username %q", ErrUserAlreadyExists, username)
	}
	if existing, err := s.users.FindByEmail(ctx, email); err != nil {
		return nil, err
	} else if existing != nil {
		return nil, fmt.Errorf("%w: email %q", ErrUserAlreadyExists, email)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Username:      username,
		Email:         email,
		PasswordHash:  string(hash),
		IsActive:      true,
		EmailVerified: false,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	// 送信失敗で登録自体は巻き戻さない(再送導線があるため)。ログに残して気づけるようにする。
	if err := sendEmailVerification(ctx, s.jwt, s.mailer, s.emailVerifyURL, user); err != nil {
		slog.Warn("failed to send verification email", "user_id", user.ID, "error", err)
	}
	return user, nil
}

func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	userID, tokenEmail, err := s.jwt.ParseEmailVerify(token)
	if err != nil {
		return ErrInvalidToken
	}
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	// 発行時のアドレスと現在のアドレスが一致するときだけ確認済みにする。メール変更後に
	// 古いリンクを踏んでも、確認していない現在のアドレスを確認済みにしないため。
	if user == nil || user.Email != tokenEmail {
		return ErrInvalidToken
	}
	return s.users.SetEmailVerified(ctx, userID, true)
}

func (s *authService) ResendVerification(ctx context.Context, email string) error {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	// 不在・確認済みは何もしない(存在有無を漏らさない)。
	if user == nil || user.EmailVerified {
		return nil
	}
	return sendEmailVerification(ctx, s.jwt, s.mailer, s.emailVerifyURL, user)
}

func (s *authService) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	// 不在でも成功と同じ挙動(メール列挙を防ぐ)。
	if user == nil {
		return nil
	}
	return sendPasswordReset(ctx, s.jwt, s.mailer, s.passwordResetURL, user)
}

func (s *authService) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	userID, err := s.jwt.Parse(token, jwtpkg.TypePasswordReset)
	if err != nil {
		return ErrInvalidToken
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.users.UpdatePassword(ctx, userID, string(hash)); err != nil {
		return err
	}
	// リセットリンクの到達自体がメール到達性の証明になるため、未確認のまま来ても確認済みにする
	//(未確認ユーザーがリセット後もログインできない事態を防ぐ)。
	return s.users.SetEmailVerified(ctx, userID, true)
}
