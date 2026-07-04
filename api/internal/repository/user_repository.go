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

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindAll(ctx context.Context) ([]domain.User, error) {
	var users []domain.User
	err := r.db.WithContext(ctx).Order("id").Find(&users).Error
	return users, err
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).
		Model(&domain.User{ID: user.ID}).
		Updates(map[string]any{"username": user.Username, "email": user.Email}).Error
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	return r.db.WithContext(ctx).
		Model(&domain.User{ID: userID}).
		Update("password_hash", passwordHash).Error
}

func (r *userRepository) UpdateAvatarKey(ctx context.Context, userID string, key *string) error {
	return r.db.WithContext(ctx).
		Model(&domain.User{ID: userID}).
		Update("avatar_key", key).Error
}

func (r *userRepository) Delete(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 所有レシピを先に消す(recipes.owner_id には CASCADE が無いため)。
		// レシピの子テーブル・共有・並び順・アーカイブはレシピ削除で CASCADE。
		if err := tx.Where("owner_id = ?", userID).Delete(&domain.Recipe{}).Error; err != nil {
			return err
		}
		// user 削除で labels / recipe_orders / recipe_archives / 共有(共有先側)は CASCADE。
		return tx.Where("id = ?", userID).Delete(&domain.User{}).Error
	})
}
