package service

import (
	"context"

	"recipe-backend/internal/domain"
)

// --- RecipeRepository のモック ---

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

// --- UserRepository のモック ---

type mockUserRepo struct {
	byName  map[string]*domain.ApplicationUser
	byEmail map[string]*domain.ApplicationUser
	byID    map[uint]*domain.ApplicationUser
	all     []domain.ApplicationUser
	nextID  uint
}

func (m *mockUserRepo) FindByUsername(_ context.Context, username string) (*domain.ApplicationUser, error) {
	return m.byName[username], nil
}
func (m *mockUserRepo) FindByEmail(_ context.Context, email string) (*domain.ApplicationUser, error) {
	return m.byEmail[email], nil
}
func (m *mockUserRepo) FindByID(_ context.Context, id uint) (*domain.ApplicationUser, error) {
	return m.byID[id], nil
}
func (m *mockUserRepo) FindAll(_ context.Context) ([]domain.ApplicationUser, error) {
	return m.all, nil
}
func (m *mockUserRepo) Create(_ context.Context, user *domain.ApplicationUser) error {
	m.nextID++
	user.ID = m.nextID
	if m.byName == nil {
		m.byName = map[string]*domain.ApplicationUser{}
	}
	if m.byEmail == nil {
		m.byEmail = map[string]*domain.ApplicationUser{}
	}
	m.byName[user.Username] = user
	m.byEmail[user.Email] = user
	return nil
}

// --- LabelRepository のモック ---

type mockLabelRepo struct {
	labels []domain.RecipeLabel
	err    error
}

func (m *mockLabelRepo) FindAll(_ context.Context) ([]domain.RecipeLabel, error) {
	return m.labels, m.err
}
