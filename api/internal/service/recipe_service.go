package service

import (
	"context"
	"fmt"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/dto/request"
)

type RecipeService interface {
	List(ctx context.Context, userID uint) ([]domain.Recipe, error)
	Create(ctx context.Context, userID uint, req request.RecipeRequest) (*domain.Recipe, error)
	Update(ctx context.Context, userID, recipeID uint, req request.RecipeRequest) (*domain.Recipe, error)
	Delete(ctx context.Context, userID, recipeID uint) error
}

type recipeService struct {
	recipes domain.RecipeRepository
	users   domain.UserRepository
}

func NewRecipeService(recipes domain.RecipeRepository, users domain.UserRepository) RecipeService {
	return &recipeService{recipes: recipes, users: users}
}

func (s *recipeService) List(ctx context.Context, userID uint) ([]domain.Recipe, error) {
	return s.recipes.FindAllForUser(ctx, userID)
}

func (s *recipeService) Create(ctx context.Context, userID uint, req request.RecipeRequest) (*domain.Recipe, error) {
	recipe := &domain.Recipe{
		Title:      req.Title,
		CreateTime: req.CreateTime,
		CreateFor:  normalizeCreateFor(req.CreateFor),
		Procedure:  req.Procedure,
		ArchiveFlg: req.ArchiveFlg,
		OwnerID:    userID,
	}
	if err := s.buildAssociations(ctx, req, recipe); err != nil {
		return nil, err
	}
	if err := s.recipes.Create(ctx, recipe); err != nil {
		return nil, err
	}
	return s.recipes.FindByID(ctx, recipe.ID)
}

func (s *recipeService) Update(ctx context.Context, userID, recipeID uint, req request.RecipeRequest) (*domain.Recipe, error) {
	existing, err := s.recipes.FindByID(ctx, recipeID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	if !canModify(existing, userID) {
		return nil, ErrForbidden
	}

	existing.Title = req.Title
	existing.CreateTime = req.CreateTime
	existing.CreateFor = normalizeCreateFor(req.CreateFor)
	existing.Procedure = req.Procedure
	existing.ArchiveFlg = req.ArchiveFlg
	// owner は変更しない
	if err := s.buildAssociations(ctx, req, existing); err != nil {
		return nil, err
	}
	if err := s.recipes.Update(ctx, existing); err != nil {
		return nil, err
	}
	return s.recipes.FindByID(ctx, recipeID)
}

func (s *recipeService) Delete(ctx context.Context, userID, recipeID uint) error {
	existing, err := s.recipes.FindByID(ctx, recipeID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrNotFound
	}
	if !canModify(existing, userID) {
		return ErrForbidden
	}
	return s.recipes.Delete(ctx, existing)
}

// buildAssociations はリクエストから label/shared_user/cooking/season を解決し recipe に詰める。
// label/ingredient/seasoning は name で get-or-create、shared_user は username 検索（無ければエラー）。
func (s *recipeService) buildAssociations(ctx context.Context, req request.RecipeRequest, recipe *domain.Recipe) error {
	labels := make([]domain.RecipeLabel, 0, len(req.Label))
	for _, l := range req.Label {
		obj, err := s.recipes.GetOrCreateLabel(ctx, l.Name)
		if err != nil {
			return err
		}
		labels = append(labels, *obj)
	}
	recipe.Labels = labels

	shared := make([]domain.ApplicationUser, 0, len(req.SharedUser))
	for _, su := range req.SharedUser {
		u, err := s.users.FindByUsername(ctx, su.Username)
		if err != nil {
			return err
		}
		if u == nil {
			return fmt.Errorf("%w: %s", ErrSharedUserNotFound, su.Username)
		}
		shared = append(shared, *u)
	}
	recipe.SharedUsers = shared

	cooking := make([]domain.Cooking, 0, len(req.Cooking))
	for _, c := range req.Cooking {
		ing, err := s.recipes.GetOrCreateIngredient(ctx, c.Ingredients.Name)
		if err != nil {
			return err
		}
		cooking = append(cooking, domain.Cooking{
			IngredientID: ing.ID,
			Quantity:     c.Quantity,
			Unit:         c.Unit,
		})
	}
	recipe.Cooking = cooking

	season := make([]domain.Season, 0, len(req.Season))
	for _, se := range req.Season {
		sea, err := s.recipes.GetOrCreateSeasoning(ctx, se.Seasoning.Name)
		if err != nil {
			return err
		}
		season = append(season, domain.Season{
			SeasoningID: sea.ID,
			Quantity:    se.Quantity,
			Unit:        se.Unit,
		})
	}
	recipe.Season = season
	return nil
}

// canModify は owner もしくは共有先ユーザーであれば true（DRF の IsDisplay 相当）。
func canModify(r *domain.Recipe, userID uint) bool {
	if r.OwnerID == userID {
		return true
	}
	for i := range r.SharedUsers {
		if r.SharedUsers[i].ID == userID {
			return true
		}
	}
	return false
}

// normalizeCreateFor は未指定(0)のとき Django のデフォルト 1 に合わせる。
func normalizeCreateFor(v int) int {
	if v == 0 {
		return 1
	}
	return v
}
