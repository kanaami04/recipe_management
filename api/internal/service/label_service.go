package service

import (
	"context"

	"recipe-backend/internal/domain"
)

type LabelService interface {
	List(ctx context.Context) ([]domain.RecipeLabel, error)
}

type labelService struct {
	labels domain.LabelRepository
}

func NewLabelService(labels domain.LabelRepository) LabelService {
	return &labelService{labels: labels}
}

func (s *labelService) List(ctx context.Context) ([]domain.RecipeLabel, error) {
	return s.labels.FindAll(ctx)
}
