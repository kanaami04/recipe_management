package service

import (
	"context"

	"recipe-backend/internal/domain"
)

type LabelService interface {
	// List は userID が閲覧できるレシピに付いたラベル名を重複なく返す。
	List(ctx context.Context, userID string) ([]string, error)
}

type labelService struct {
	labels domain.LabelRepository
}

func NewLabelService(labels domain.LabelRepository) LabelService {
	return &labelService{labels: labels}
}

func (s *labelService) List(ctx context.Context, userID string) ([]string, error) {
	return s.labels.FindNamesForUser(ctx, userID)
}
