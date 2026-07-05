package service

import (
	"context"
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
	groups  domain.ShareGroupRepository
}

func NewRecipeService(recipes domain.RecipeRepository, groups domain.ShareGroupRepository) RecipeService {
	return &recipeService{recipes: recipes, groups: groups}
}

func (s *recipeService) List(ctx context.Context, userID string) ([]domain.Recipe, error) {
	recipes, err := s.recipes.FindAllForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.fillSharedUsers(ctx, userID, recipes); err != nil {
		return nil, err
	}
	return recipes, nil
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
	buildAssociations(req, recipe)
	if err := s.recipes.Create(ctx, recipe); err != nil {
		return nil, err
	}
	created, err := s.recipes.FindByID(ctx, recipe.ID)
	if err != nil {
		return nil, err
	}
	if err := s.fillSharedUsersOne(ctx, userID, created); err != nil {
		return nil, err
	}
	return created, nil
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
	if ok, err := s.canModify(ctx, existing, userID); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrForbidden
	}

	existing.Title = req.Title
	existing.CookingTime = req.CreateTime
	existing.Servings = normalizeServings(req.CreateFor)
	existing.Procedure = req.Procedure
	existing.SourceURL = req.SourceUrl
	existing.ThumbnailURL = req.ThumbnailUrl
	// owner・アーカイブ状態は更新対象外(アーカイブは専用の SetArchived で扱う)
	buildAssociations(req, existing)
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
	if err := s.fillSharedUsersOne(ctx, userID, updated); err != nil {
		return nil, err
	}
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
	if ok, err := s.canModify(ctx, existing, userID); err != nil {
		return err
	} else if !ok {
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
	if ok, err := s.canModify(ctx, existing, userID); err != nil {
		return err
	} else if !ok {
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

// buildAssociations はリクエストから label/cooking/season を解決し recipe に詰める。
// いずれもレシピ従属の子レコードとして名前をそのまま持つ。共有はシェアグループで管理するため
// ここでは扱わない。
func buildAssociations(req request.RecipeRequest, recipe *domain.Recipe) {
	labels := make([]domain.RecipeLabel, 0, len(req.Label))
	for _, l := range req.Label {
		labels = append(labels, domain.RecipeLabel{Name: l.Name})
	}
	recipe.Labels = labels

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
}

// canModify は userID が recipe を編集できる(所有者、または owner と同じシェアグループの
// メンバー)かを返す。グループ内は所有物を全員で共同編集できる。
func (s *recipeService) canModify(ctx context.Context, r *domain.Recipe, userID string) (bool, error) {
	if r.OwnerID == userID {
		return true, nil
	}
	memberIDs, err := s.groups.MemberIDs(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, id := range memberIDs {
		if id == r.OwnerID {
			return true, nil
		}
	}
	return false, nil
}

// fillSharedUsers は各レシピの SharedUsers に「操作ユーザーのシェアグループのメンバー
// (各レシピの owner を除く)」を詰める。グループ未所属なら空のまま。
func (s *recipeService) fillSharedUsers(ctx context.Context, userID string, recipes []domain.Recipe) error {
	group, err := s.groups.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if group == nil {
		return nil
	}
	for i := range recipes {
		recipes[i].SharedUsers = membersExcept(group.Members, recipes[i].OwnerID)
	}
	return nil
}

// fillSharedUsersOne は 1 件のレシピについて fillSharedUsers と同じ処理を行う。
func (s *recipeService) fillSharedUsersOne(ctx context.Context, userID string, recipe *domain.Recipe) error {
	group, err := s.groups.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if group == nil {
		return nil
	}
	recipe.SharedUsers = membersExcept(group.Members, recipe.OwnerID)
	return nil
}

// membersExcept は members から ownerID のユーザーを除いたスライスを返す。
func membersExcept(members []domain.User, ownerID string) []domain.User {
	out := make([]domain.User, 0, len(members))
	for i := range members {
		if members[i].ID != ownerID {
			out = append(out, members[i])
		}
	}
	return out
}

// normalizeServings は未指定(0)のとき既定の 1 人前に揃える。
func normalizeServings(v int) int {
	if v == 0 {
		return 1
	}
	return v
}
