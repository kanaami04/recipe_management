package service

import (
	"context"
	"errors"
	"testing"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/dto/request"
)

// --- モック実装 ---

type mockRecipeRepo struct {
	store      map[uint]*domain.Recipe
	nextID     uint
	labelSeq   uint
	ingSeq     uint
	seaSeq     uint
	created    *domain.Recipe
	updated    *domain.Recipe
	deletedIDs []uint
}

func newMockRecipeRepo() *mockRecipeRepo {
	return &mockRecipeRepo{store: map[uint]*domain.Recipe{}, nextID: 1}
}

func (m *mockRecipeRepo) FindAllForUser(_ context.Context, userID uint) ([]domain.Recipe, error) {
	var out []domain.Recipe
	for _, r := range m.store {
		out = append(out, *r)
	}
	return out, nil
}
func (m *mockRecipeRepo) FindByID(_ context.Context, id uint) (*domain.Recipe, error) {
	r, ok := m.store[id]
	if !ok {
		return nil, nil
	}
	return r, nil
}
func (m *mockRecipeRepo) Create(_ context.Context, recipe *domain.Recipe) error {
	recipe.ID = m.nextID
	m.nextID++
	cp := *recipe
	m.store[recipe.ID] = &cp
	m.created = &cp
	return nil
}
func (m *mockRecipeRepo) Update(_ context.Context, recipe *domain.Recipe) error {
	cp := *recipe
	m.store[recipe.ID] = &cp
	m.updated = &cp
	return nil
}
func (m *mockRecipeRepo) Delete(_ context.Context, recipe *domain.Recipe) error {
	delete(m.store, recipe.ID)
	m.deletedIDs = append(m.deletedIDs, recipe.ID)
	return nil
}
func (m *mockRecipeRepo) GetOrCreateLabel(_ context.Context, name string) (*domain.RecipeLabel, error) {
	m.labelSeq++
	return &domain.RecipeLabel{ID: m.labelSeq, Name: name}, nil
}
func (m *mockRecipeRepo) GetOrCreateIngredient(_ context.Context, name string) (*domain.Ingredient, error) {
	m.ingSeq++
	return &domain.Ingredient{ID: m.ingSeq, Name: name}, nil
}
func (m *mockRecipeRepo) GetOrCreateSeasoning(_ context.Context, name string) (*domain.Seasoning, error) {
	m.seaSeq++
	return &domain.Seasoning{ID: m.seaSeq, Name: name}, nil
}

type mockUserRepo struct {
	byName map[string]*domain.ApplicationUser
}

func (m *mockUserRepo) FindByUsername(_ context.Context, username string) (*domain.ApplicationUser, error) {
	u, ok := m.byName[username]
	if !ok {
		return nil, nil
	}
	return u, nil
}
func (m *mockUserRepo) FindByID(_ context.Context, id uint) (*domain.ApplicationUser, error) {
	return nil, nil
}
func (m *mockUserRepo) FindAll(_ context.Context) ([]domain.ApplicationUser, error)  { return nil, nil }
func (m *mockUserRepo) Create(_ context.Context, user *domain.ApplicationUser) error { return nil }

// --- テスト ---

func TestCreate_SetsOwnerAndDefaults(t *testing.T) {
	rr := newMockRecipeRepo()
	ur := &mockUserRepo{byName: map[string]*domain.ApplicationUser{}}
	svc := NewRecipeService(rr, ur)

	_, err := svc.Create(context.Background(), 42, request.RecipeRequest{
		Title: "カレー",
		// create_for 未指定 → 1 に正規化される想定
		Label:   []request.LabelInput{{Name: "夕食"}},
		Cooking: []request.CookingInput{{Ingredients: request.NameInput{Name: "玉ねぎ"}, Quantity: 2, Unit: "個"}},
		Season:  []request.SeasonInput{{Seasoning: request.NameInput{Name: "塩"}, Quantity: 1, Unit: "g"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rr.created.OwnerID != 42 {
		t.Errorf("owner = %d, want 42", rr.created.OwnerID)
	}
	if rr.created.CreateFor != 1 {
		t.Errorf("create_for = %d, want 1 (default)", rr.created.CreateFor)
	}
	if len(rr.created.Labels) != 1 || len(rr.created.Cooking) != 1 || len(rr.created.Season) != 1 {
		t.Errorf("associations not built: labels=%d cooking=%d season=%d",
			len(rr.created.Labels), len(rr.created.Cooking), len(rr.created.Season))
	}
}

func TestCreate_SharedUserNotFound(t *testing.T) {
	rr := newMockRecipeRepo()
	ur := &mockUserRepo{byName: map[string]*domain.ApplicationUser{}}
	svc := NewRecipeService(rr, ur)

	_, err := svc.Create(context.Background(), 1, request.RecipeRequest{
		Title:      "親子丼",
		SharedUser: []request.SharedUserInput{{Username: "ghost"}},
	})
	if !errors.Is(err, ErrSharedUserNotFound) {
		t.Fatalf("err = %v, want ErrSharedUserNotFound", err)
	}
}

func TestUpdate_ForbiddenForNonOwnerNonShared(t *testing.T) {
	rr := newMockRecipeRepo()
	rr.store[1] = &domain.Recipe{ID: 1, OwnerID: 100} // 所有者は 100
	ur := &mockUserRepo{byName: map[string]*domain.ApplicationUser{}}
	svc := NewRecipeService(rr, ur)

	_, err := svc.Update(context.Background(), 999, 1, request.RecipeRequest{Title: "x"}) // 別人 999
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
}

func TestUpdate_AllowedForSharedUser(t *testing.T) {
	rr := newMockRecipeRepo()
	rr.store[1] = &domain.Recipe{
		ID:          1,
		OwnerID:     100,
		SharedUsers: []domain.ApplicationUser{{ID: 7}},
	}
	ur := &mockUserRepo{byName: map[string]*domain.ApplicationUser{}}
	svc := NewRecipeService(rr, ur)

	_, err := svc.Update(context.Background(), 7, 1, request.RecipeRequest{Title: "更新"}) // 共有先 7 は許可
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rr.updated == nil || rr.updated.Title != "更新" {
		t.Errorf("update not applied: %+v", rr.updated)
	}
	if rr.updated.OwnerID != 100 {
		t.Errorf("owner changed to %d, want unchanged 100", rr.updated.OwnerID)
	}
}

func TestDelete_NotFound(t *testing.T) {
	rr := newMockRecipeRepo()
	ur := &mockUserRepo{byName: map[string]*domain.ApplicationUser{}}
	svc := NewRecipeService(rr, ur)

	err := svc.Delete(context.Background(), 1, 12345)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
