package service

import (
	"context"
	"fmt"

	"recipe-backend/internal/domain"
	jwtpkg "recipe-backend/internal/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (access, refresh string, err error)
	Refresh(ctx context.Context, refreshToken string) (access string, err error)
	Register(ctx context.Context, username, email, password string) (*domain.ApplicationUser, error)
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

// Register は新規ユーザーを作成する。username/email の重複は ErrUserAlreadyExists。
func (s *authService) Register(ctx context.Context, username, email, password string) (*domain.ApplicationUser, error) {
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

	user := &domain.ApplicationUser{
		Username: username,
		Email:    email,
		Password: string(hash),
		IsActive: true,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}
