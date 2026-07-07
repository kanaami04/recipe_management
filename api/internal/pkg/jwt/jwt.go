package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTTL        = time.Hour          // アクセストークン: 1時間
	refreshTTL       = 7 * 24 * time.Hour // リフレッシュトークン: 7日
	emailVerifyTTL   = 24 * time.Hour     // メール確認トークン: 24時間
	passwordResetTTL = time.Hour          // パスワードリセットトークン: 1時間

	TypeAccess        = "access"
	TypeRefresh       = "refresh"
	TypeEmailVerify   = "email_verify"
	TypePasswordReset = "password_reset"
)

// Manager は JWT の発行・検証を行う（simplejwt 互換の最小実装）。
type Manager struct {
	secret []byte
}

func NewManager(secret string) *Manager {
	return &Manager{secret: []byte(secret)}
}

func (m *Manager) generate(userID string, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id":    userID,
		"token_type": tokenType,
		"exp":        now.Add(ttl).Unix(),
		"iat":        now.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) GenerateAccess(userID string) (string, error) {
	return m.generate(userID, TypeAccess, accessTTL)
}

func (m *Manager) GenerateRefresh(userID string) (string, error) {
	return m.generate(userID, TypeRefresh, refreshTTL)
}

// GenerateEmailVerify はメール確認用の署名付きトークン(有効期限 24h)を発行する。
// DB に確認用カラムを増やさず、Parse で token_type と有効期限を検証する。
// email を claim に埋め、検証時に「発行時のアドレス == 現在のアドレス」を確かめられるようにする
// (メール変更後に古いリンクで別アドレスを確認済みにしないため)。
func (m *Manager) GenerateEmailVerify(userID, email string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id":    userID,
		"email":      email,
		"token_type": TypeEmailVerify,
		"exp":        now.Add(emailVerifyTTL).Unix(),
		"iat":        now.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// GeneratePasswordReset はパスワードリセット用の署名付きトークン(有効期限 1h)を発行する。
func (m *Manager) GeneratePasswordReset(userID string) (string, error) {
	return m.generate(userID, TypePasswordReset, passwordResetTTL)
}

// parseClaims はトークンの署名・有効期限・token_type を検証し、claims を返す。
func (m *Manager) parseClaims(tokenString, expectedType string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	if tt, _ := claims["token_type"].(string); tt != expectedType {
		return nil, errors.New("invalid token type")
	}
	return claims, nil
}

// Parse はトークンを検証し、token_type が expectedType と一致すれば user_id を返す。
func (m *Manager) Parse(tokenString, expectedType string) (string, error) {
	claims, err := m.parseClaims(tokenString, expectedType)
	if err != nil {
		return "", err
	}
	uid, ok := claims["user_id"].(string)
	if !ok || uid == "" {
		return "", errors.New("invalid user_id")
	}
	return uid, nil
}

// ParseEmailVerify はメール確認トークンを検証し、user_id と埋め込まれた email を返す。
// 呼び出し側はこの email が現在のアカウントのアドレスと一致するかを確かめる。
func (m *Manager) ParseEmailVerify(tokenString string) (userID, email string, err error) {
	claims, err := m.parseClaims(tokenString, TypeEmailVerify)
	if err != nil {
		return "", "", err
	}
	uid, _ := claims["user_id"].(string)
	em, _ := claims["email"].(string)
	if uid == "" {
		return "", "", errors.New("invalid user_id")
	}
	return uid, em, nil
}
