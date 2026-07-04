package service

import (
	"context"

	"recipe-backend/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	// UpdateProfile は username を更新する。他ユーザーと重複したら ErrUserAlreadyExists。
	// メールは本人確認が要るため ChangeEmail で別途扱う。
	UpdateProfile(ctx context.Context, userID, username string) (*domain.User, error)
	// ChangeEmail は現在のパスワードを検証して email を変える。
	// パスワードが違えば ErrIncorrectPassword、他ユーザーと重複したら ErrUserAlreadyExists。
	ChangeEmail(ctx context.Context, userID, email, password string) (*domain.User, error)
	// ChangePassword は現在のパスワードを検証して新しいものに変える。
	// 現在のパスワードが違えば ErrIncorrectPassword。
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
	// DeleteAccount はユーザーと所有レシピを削除する。
	DeleteAccount(ctx context.Context, userID string) error
}

type userService struct {
	users domain.UserRepository
}

func NewUserService(users domain.UserRepository) UserService {
	return &userService{users: users}
}

func (s *userService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return s.users.FindByID(ctx, id)
}

func (s *userService) List(ctx context.Context) ([]domain.User, error) {
	return s.users.FindAll(ctx)
}

func (s *userService) UpdateProfile(ctx context.Context, userID, username string) (*domain.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}
	// 変更するときだけ、他ユーザーとの重複を確かめる。
	if username != user.Username {
		if err := s.ensureUsernameFree(ctx, username, userID); err != nil {
			return nil, err
		}
	}
	user.Username = username
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) ChangeEmail(ctx context.Context, userID, email, password string) (*domain.User, error) {
	user, err := s.authenticate(ctx, userID, password)
	if err != nil {
		return nil, err
	}
	if email != user.Email {
		if err := s.ensureEmailFree(ctx, email, userID); err != nil {
			return nil, err
		}
	}
	user.Email = email
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	if _, err := s.authenticate(ctx, userID, currentPassword); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.users.UpdatePassword(ctx, userID, string(hash))
}

func (s *userService) DeleteAccount(ctx context.Context, userID string) error {
	return s.users.Delete(ctx, userID)
}

// authenticate は userID のユーザーを取得し、現在のパスワードを検証する。
// ChangeEmail / ChangePassword で共通の「本人確認」に使う。
func (s *userService) authenticate(ctx context.Context, userID, password string) (*domain.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, ErrIncorrectPassword
	}
	return user, nil
}

// ensureUsernameFree は username が自分以外に使われていないことを確かめる。
func (s *userService) ensureUsernameFree(ctx context.Context, username, selfID string) error {
	existing, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != selfID {
		return ErrUserAlreadyExists
	}
	return nil
}

// ensureEmailFree は email が自分以外に使われていないことを確かめる。
func (s *userService) ensureEmailFree(ctx context.Context, email, selfID string) error {
	existing, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != selfID {
		return ErrUserAlreadyExists
	}
	return nil
}
