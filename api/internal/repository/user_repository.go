package repository

import (
	"context"
	"errors"

	"recipe-backend/internal/domain"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*domain.ApplicationUser, error) {
	var u domain.ApplicationUser
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*domain.ApplicationUser, error) {
	var u domain.ApplicationUser
	err := r.db.WithContext(ctx).First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindAll(ctx context.Context) ([]domain.ApplicationUser, error) {
	var users []domain.ApplicationUser
	err := r.db.WithContext(ctx).Order("id").Find(&users).Error
	return users, err
}

func (r *userRepository) Create(ctx context.Context, user *domain.ApplicationUser) error {
	return r.db.WithContext(ctx).Create(user).Error
}
