package service

import (
	"context"

	"recipe-backend/internal/domain"
)

type LabelService interface {
	// List は ownerID が管理するラベルを返す。
	List(ctx context.Context, ownerID string) ([]domain.Label, error)
	// Create はラベルを作る。同名が既にあれば ErrDuplicate。
	Create(ctx context.Context, ownerID, name string) (*domain.Label, error)
	// Rename はラベルを改名する。未所有は ErrForbidden、無ければ ErrNotFound、
	// 改名先が既にあれば ErrDuplicate。
	Rename(ctx context.Context, ownerID, id, name string) (*domain.Label, error)
	// Delete はラベルを削除する。未所有は ErrForbidden、無ければ ErrNotFound。
	Delete(ctx context.Context, ownerID, id string) error
}

type labelService struct {
	labels domain.LabelRepository
}

func NewLabelService(labels domain.LabelRepository) LabelService {
	return &labelService{labels: labels}
}

func (s *labelService) List(ctx context.Context, ownerID string) ([]domain.Label, error) {
	return s.labels.FindAllForOwner(ctx, ownerID)
}

func (s *labelService) Create(ctx context.Context, ownerID, name string) (*domain.Label, error) {
	existing, err := s.labels.FindByOwnerAndName(ctx, ownerID, name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDuplicate
	}
	label := &domain.Label{Name: name, OwnerID: ownerID}
	if err := s.labels.Create(ctx, label); err != nil {
		return nil, err
	}
	return label, nil
}

func (s *labelService) Rename(ctx context.Context, ownerID, id, name string) (*domain.Label, error) {
	label, err := s.requireOwned(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}
	if name == label.Name {
		return label, nil // 変化なし
	}
	existing, err := s.labels.FindByOwnerAndName(ctx, ownerID, name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDuplicate
	}
	if err := s.labels.Rename(ctx, label, name); err != nil {
		return nil, err
	}
	label.Name = name
	return label, nil
}

func (s *labelService) Delete(ctx context.Context, ownerID, id string) error {
	label, err := s.requireOwned(ctx, ownerID, id)
	if err != nil {
		return err
	}
	return s.labels.Delete(ctx, label)
}

// requireOwned は id のラベルを取得し、存在と所有者一致を確かめて返す。
func (s *labelService) requireOwned(ctx context.Context, ownerID, id string) (*domain.Label, error) {
	label, err := s.labels.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if label == nil {
		return nil, ErrNotFound
	}
	if label.OwnerID != ownerID {
		return nil, ErrForbidden
	}
	return label, nil
}
