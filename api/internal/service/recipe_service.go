package service

import (
	"context"
	"fmt"
	"net/url"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/dto/request"
)

type RecipeService interface {
	List(ctx context.Context, userID string) ([]domain.Recipe, error)
	Create(ctx context.Context, userID string, req request.RecipeRequest) (*domain.Recipe, error)
	Update(ctx context.Context, userID, recipeID string, req request.RecipeRequest) (*domain.Recipe, error)
	Delete(ctx context.Context, userID, recipeID string) error
	Reorder(ctx context.Context, userID string, recipeIDs []string) error
	SetArchived(ctx context.Context, userID, recipeID string, archived bool) error
}

type recipeService struct {
	recipes domain.RecipeRepository
	users   domain.UserRepository
}

func NewRecipeService(recipes domain.RecipeRepository, users domain.UserRepository) RecipeService {
	return &recipeService{recipes: recipes, users: users}
}

func (s *recipeService) List(ctx context.Context, userID string) ([]domain.Recipe, error) {
	return s.recipes.FindAllForUser(ctx, userID)
}

// validateExternalURLs は source_url / thumbnail_url を検証する。
// 空文字は許容し、値があるときは http/https のスキームのみ通す
// (javascript: 等の危険な URL が保存され <a href>・<img src> で描画されるのを防ぐ)。
func validateExternalURLs(req request.RecipeRequest) error {
	for _, raw := range []string{req.SourceUrl, req.ThumbnailUrl} {
		if raw == "" {
			continue
		}
		u, err := url.Parse(raw)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
			return ErrInvalidURL
		}
	}
	return nil
}

func (s *recipeService) Create(ctx context.Context, userID string, req request.RecipeRequest) (*domain.Recipe, error) {
	if err := validateExternalURLs(req); err != nil {
		return nil, err
	}
	recipe := &domain.Recipe{
		Title:        req.Title,
		CookingTime:  req.CreateTime,
		Servings:     normalizeServings(req.CreateFor),
		Procedure:    req.Procedure,
		SourceURL:    req.SourceUrl,
		ThumbnailURL: req.ThumbnailUrl,
		OwnerID:      userID,
	}
	if err := s.buildAssociations(ctx, req, recipe); err != nil {
		return nil, err
	}
	if err := s.recipes.Create(ctx, recipe); err != nil {
		return nil, err
	}
	return s.recipes.FindByID(ctx, recipe.ID)
}

func (s *recipeService) Update(ctx context.Context, userID, recipeID string, req request.RecipeRequest) (*domain.Recipe, error) {
	if err := validateExternalURLs(req); err != nil {
		return nil, err
	}
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
	existing.CookingTime = req.CreateTime
	existing.Servings = normalizeServings(req.CreateFor)
	existing.Procedure = req.Procedure
	existing.SourceURL = req.SourceUrl
	existing.ThumbnailURL = req.ThumbnailUrl
	// owner・アーカイブ状態は更新対象外(アーカイブは専用の SetArchived で扱う)
	if err := s.buildAssociations(ctx, req, existing); err != nil {
		return nil, err
	}
	if err := s.recipes.Update(ctx, existing); err != nil {
		return nil, err
	}
	updated, err := s.recipes.FindByID(ctx, recipeID)
	if err != nil {
		return nil, err
	}
	// レスポンスの archive_flg は「操作したユーザーにとっての状態」を返す。
	archived, err := s.recipes.IsArchived(ctx, userID, recipeID)
	if err != nil {
		return nil, err
	}
	updated.Archived = archived
	return updated, nil
}

// SetArchived は userID にとっての recipeID のアーカイブ状態を切り替える。
// 閲覧できる(所有 or 共有された)レシピにのみ許可し、他ユーザーの状態には影響しない。
func (s *recipeService) SetArchived(ctx context.Context, userID, recipeID string, archived bool) error {
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
	return s.recipes.SetArchived(ctx, userID, recipeID, archived)
}

func (s *recipeService) Delete(ctx context.Context, userID, recipeID string) error {
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

// Reorder は userID の一覧表示順を recipeIDs の並びで保存する。
// 指定 ID は全て「そのユーザーが閲覧できる(所有 or 共有された)レシピ」でなければ
// ならない。見えないレシピの順序を書こうとした場合は ErrForbidden。
func (s *recipeService) Reorder(ctx context.Context, userID string, recipeIDs []string) error {
	visible, err := s.recipes.FindAllForUser(ctx, userID)
	if err != nil {
		return err
	}
	allowed := make(map[string]struct{}, len(visible))
	for i := range visible {
		allowed[visible[i].ID] = struct{}{}
	}
	// 閲覧不可を弾きつつ重複を除く(最初の出現順を維持)。重複を残すと
	// upsert が同一 (user, recipe) を二重に対象にして DB エラーになる。
	seen := make(map[string]struct{}, len(recipeIDs))
	deduped := make([]string, 0, len(recipeIDs))
	for _, rid := range recipeIDs {
		if _, ok := allowed[rid]; !ok {
			return ErrForbidden
		}
		if _, dup := seen[rid]; dup {
			continue
		}
		seen[rid] = struct{}{}
		deduped = append(deduped, rid)
	}
	return s.recipes.Reorder(ctx, userID, deduped)
}

// buildAssociations はリクエストから label/shared_user/cooking/season を解決し recipe に詰める。
// label/ingredient/seasoning はレシピ従属の子レコードとして名前をそのまま持つ。
// shared_user のみ username でユーザーを検索する（無ければエラー）。
func (s *recipeService) buildAssociations(ctx context.Context, req request.RecipeRequest, recipe *domain.Recipe) error {
	labels := make([]domain.RecipeLabel, 0, len(req.Label))
	for _, l := range req.Label {
		labels = append(labels, domain.RecipeLabel{Name: l.Name})
	}
	recipe.Labels = labels

	shared := make([]domain.User, 0, len(req.SharedUser))
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

	ingredients := make([]domain.RecipeIngredient, 0, len(req.Cooking))
	for _, c := range req.Cooking {
		ingredients = append(ingredients, domain.RecipeIngredient{
			Name:     c.Ingredients.Name,
			Quantity: c.Quantity,
			Unit:     c.Unit,
		})
	}
	recipe.Ingredients = ingredients

	seasonings := make([]domain.RecipeSeasoning, 0, len(req.Season))
	for _, se := range req.Season {
		seasonings = append(seasonings, domain.RecipeSeasoning{
			Name:     se.Seasoning.Name,
			Quantity: se.Quantity,
			Unit:     se.Unit,
		})
	}
	recipe.Seasonings = seasonings
	return nil
}

// canModify は owner もしくは共有先ユーザーであれば true（DRF の IsDisplay 相当）。
func canModify(r *domain.Recipe, userID string) bool {
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

// normalizeServings は未指定(0)のとき既定の 1 人前に揃える。
func normalizeServings(v int) int {
	if v == 0 {
		return 1
	}
	return v
}
