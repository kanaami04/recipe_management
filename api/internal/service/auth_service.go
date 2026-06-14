package service

import (
	"context"

	"recipe-backend/internal/domain"
	jwtpkg "recipe-backend/internal/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (access, refresh string, err error)
	Refresh(ctx context.Context, refreshToken string) (access string, err error)
}

type authService struct {
	users domain.UserRepository
	jwt   *jwtpkg.Manager
}

func NewAuthService(users domain.UserRepository, jwt *jwtpkg.Manager) AuthService {
	return &authService{users: users, jwt: jwt}
}

func (s *authService) Login(ctx context.Context, username, password string) (string, string, error) {
	user, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		return "", "", err
	}
	if user == nil || !user.IsActive {
		return "", "", ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return "", "", ErrInvalidCredentials
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
