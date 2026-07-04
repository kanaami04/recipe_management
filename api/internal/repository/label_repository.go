package repository

import (
	"context"
	"errors"

	"recipe-backend/internal/domain"

	"gorm.io/gorm"
)

type labelRepository struct {
	db *gorm.DB
}

func NewLabelRepository(db *gorm.DB) domain.LabelRepository {
	return &labelRepository{db: db}
}

func (r *labelRepository) FindAllForOwner(ctx context.Context, ownerID string) ([]domain.Label, error) {
	var labels []domain.Label
	err := r.db.WithContext(ctx).
		Where("owner_id = ?", ownerID).
		Order("name").
		Find(&labels).Error
	return labels, err
}

func (r *labelRepository) FindByID(ctx context.Context, id string) (*domain.Label, error) {
	var label domain.Label
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&label).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &label, nil
}

func (r *labelRepository) FindByOwnerAndName(ctx context.Context, ownerID, name string) (*domain.Label, error) {
	var label domain.Label
	err := r.db.WithContext(ctx).Where("owner_id = ? AND name = ?", ownerID, name).First(&label).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &label, nil
}

func (r *labelRepository) Create(ctx context.Context, label *domain.Label) error {
	// belongs-to の Owner は FK 定義用で、書き込みでは巻き込まない。
	return r.db.WithContext(ctx).Omit("Owner").Create(label).Error
}

func (r *labelRepository) Rename(ctx context.Context, label *domain.Label, newName string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domain.Label{}).
			Where("id = ?", label.ID).
			Update("name", newName).Error; err != nil {
			return err
		}
		// 所有者のレシピの recipe_labels へ改名を伝播する。同じレシピに既に新名がある場合は
		// (recipe, name) 一意制約に触れるため、旧名の行を先に消してから残りを改名する。
		if err := tx.Exec(
			`DELETE FROM recipe_labels
			 WHERE name = ?
			   AND recipe_id IN (SELECT id FROM recipes WHERE owner_id = ?)
			   AND recipe_id IN (SELECT recipe_id FROM recipe_labels WHERE name = ?)`,
			label.Name, label.OwnerID, newName,
		).Error; err != nil {
			return err
		}
		return tx.Exec(
			`UPDATE recipe_labels SET name = ?
			 WHERE name = ? AND recipe_id IN (SELECT id FROM recipes WHERE owner_id = ?)`,
			newName, label.Name, label.OwnerID,
		).Error
	})
}

func (r *labelRepository) Delete(ctx context.Context, label *domain.Label) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 所有者のレシピからも同名ラベルを外す。
		if err := tx.Exec(
			`DELETE FROM recipe_labels
			 WHERE name = ? AND recipe_id IN (SELECT id FROM recipes WHERE owner_id = ?)`,
			label.Name, label.OwnerID,
		).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", label.ID).Delete(&domain.Label{}).Error
	})
}
