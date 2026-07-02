package service

import (
	"context"

	"recipe-backend/internal/domain"
)

// --- RecipeRepository のモック ---

type mockRecipeRepo struct {
	store      map[uint]*domain.Recipe
	nextID     uint
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
// --- UserRepository のモック ---

type mockUserRepo struct {
	byName  map[string]*domain.User
	byEmail map[string]*domain.User
	byID    map[uint]*domain.User
	all     []domain.User
	nextID  uint
}

func (m *mockUserRepo) FindByUsername(_ context.Context, username string) (*domain.User, error) {
	return m.byName[username], nil
}
func (m *mockUserRepo) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	return m.byEmail[email], nil
}
func (m *mockUserRepo) FindByID(_ context.Context, id uint) (*domain.User, error) {
	return m.byID[id], nil
}
func (m *mockUserRepo) FindAll(_ context.Context) ([]domain.User, error) {
	return m.all, nil
}
func (m *mockUserRepo) Create(_ context.Context, user *domain.User) error {
	m.nextID++
	user.ID = m.nextID
	if m.byName == nil {
		m.byName = map[string]*domain.User{}
	}
	if m.byEmail == nil {
		m.byEmail = map[string]*domain.User{}
	}
	m.byName[user.Username] = user
	m.byEmail[user.Email] = user
	return nil
}

// --- LabelRepository のモック ---

type mockLabelRepo struct {
	names []string
	err   error
}

func (m *mockLabelRepo) FindNamesForUser(_ context.Context, _ uint) ([]string, error) {
	return m.names, m.err
}
