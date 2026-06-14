package service

import (
	"context"

	"recipe-backend/internal/domain"
)

type UserService interface {
	GetByID(ctx context.Context, id uint) (*domain.ApplicationUser, error)
	List(ctx context.Context) ([]domain.ApplicationUser, error)
}

type userService struct {
	users domain.UserRepository
}

func NewUserService(users domain.UserRepository) UserService {
	return &userService{users: users}
}

func (s *userService) GetByID(ctx context.Context, id uint) (*domain.ApplicationUser, error) {
	return s.users.FindByID(ctx, id)
}

func (s *userService) List(ctx context.Context) ([]domain.ApplicationUser, error) {
	return s.users.FindAll(ctx)
}
