package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTTL  = time.Hour          // アクセストークン: 1時間
	refreshTTL = 7 * 24 * time.Hour // リフレッシュトークン: 7日

	TypeAccess  = "access"
	TypeRefresh = "refresh"
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

// Parse はトークンを検証し、token_type が expectedType と一致すれば user_id を返す。
func (m *Manager) Parse(tokenString, expectedType string) (string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}
	if tt, _ := claims["token_type"].(string); tt != expectedType {
		return "", errors.New("invalid token type")
	}
	uid, ok := claims["user_id"].(string)
	if !ok || uid == "" {
		return "", errors.New("invalid user_id")
	}
	return uid, nil
}
