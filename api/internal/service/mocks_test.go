package service

import (
	"context"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/pkg/id"
)

// --- RecipeRepository のモック ---

type mockRecipeRepo struct {
	store        map[string]*domain.Recipe
	created      *domain.Recipe
	updated      *domain.Recipe
	deletedIDs   []string
	reorderedIDs []string
	// archived[userID][recipeID] = true でアーカイブ済み。
	archived map[string]map[string]bool
}

func newMockRecipeRepo() *mockRecipeRepo {
	return &mockRecipeRepo{store: map[string]*domain.Recipe{}}
}

func (m *mockRecipeRepo) FindAllForUser(_ context.Context, _ string) ([]domain.Recipe, error) {
	var out []domain.Recipe
	for _, r := range m.store {
		out = append(out, *r)
	}
	return out, nil
}
func (m *mockRecipeRepo) FindByID(_ context.Context, id string) (*domain.Recipe, error) {
	r, ok := m.store[id]
	if !ok {
		return nil, nil
	}
	return r, nil
}
func (m *mockRecipeRepo) Create(_ context.Context, recipe *domain.Recipe) error {
	if recipe.ID == "" {
		recipe.ID = id.New()
	}
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
func (m *mockRecipeRepo) Reorder(_ context.Context, _ string, recipeIDs []string) error {
	m.reorderedIDs = recipeIDs
	return nil
}
func (m *mockRecipeRepo) SetArchived(_ context.Context, userID, recipeID string, archived bool) error {
	if m.archived == nil {
		m.archived = map[string]map[string]bool{}
	}
	if m.archived[userID] == nil {
		m.archived[userID] = map[string]bool{}
	}
	m.archived[userID][recipeID] = archived
	return nil
}
func (m *mockRecipeRepo) IsArchived(_ context.Context, userID, recipeID string) (bool, error) {
	return m.archived[userID][recipeID], nil
}

// --- UserRepository のモック ---

type mockUserRepo struct {
	byName            map[string]*domain.User
	byEmail           map[string]*domain.User
	byID              map[string]*domain.User
	all               []domain.User
	updated           *domain.User
	passwordChangedID string
	newPasswordHash   string
	deletedUserID     string
}

func (m *mockUserRepo) FindByUsername(_ context.Context, username string) (*domain.User, error) {
	return m.byName[username], nil
}
func (m *mockUserRepo) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	return m.byEmail[email], nil
}
func (m *mockUserRepo) FindByID(_ context.Context, id string) (*domain.User, error) {
	return m.byID[id], nil
}
func (m *mockUserRepo) FindAll(_ context.Context) ([]domain.User, error) {
	return m.all, nil
}
func (m *mockUserRepo) Create(_ context.Context, user *domain.User) error {
	if user.ID == "" {
		user.ID = id.New()
	}
	if m.byName == nil {
		m.byName = map[string]*domain.User{}
	}
	if m.byEmail == nil {
		m.byEmail = map[string]*domain.User{}
	}
	m.byName[user.Username] = user
	m.byEmail[user.Email] = user
	if m.byID == nil {
		m.byID = map[string]*domain.User{}
	}
	m.byID[user.ID] = user
	return nil
}
func (m *mockUserRepo) Update(_ context.Context, user *domain.User) error {
	m.updated = user
	return nil
}
func (m *mockUserRepo) UpdatePassword(_ context.Context, userID, passwordHash string) error {
	m.passwordChangedID = userID
	m.newPasswordHash = passwordHash
	return nil
}
func (m *mockUserRepo) Delete(_ context.Context, userID string) error {
	m.deletedUserID = userID
	return nil
}

// --- LabelRepository のモック ---

type mockLabelRepo struct {
	store      map[string]*domain.Label // id -> label
	created    *domain.Label
	renamedTo  string
	deletedIDs []string
	err        error
}

func newMockLabelRepo() *mockLabelRepo {
	return &mockLabelRepo{store: map[string]*domain.Label{}}
}

func (m *mockLabelRepo) FindAllForOwner(_ context.Context, ownerID string) ([]domain.Label, error) {
	if m.err != nil {
		return nil, m.err
	}
	var out []domain.Label
	for _, l := range m.store {
		if l.OwnerID == ownerID {
			out = append(out, *l)
		}
	}
	return out, nil
}
func (m *mockLabelRepo) FindByID(_ context.Context, id string) (*domain.Label, error) {
	l, ok := m.store[id]
	if !ok {
		return nil, nil
	}
	return l, nil
}
func (m *mockLabelRepo) FindByOwnerAndName(_ context.Context, ownerID, name string) (*domain.Label, error) {
	for _, l := range m.store {
		if l.OwnerID == ownerID && l.Name == name {
			return l, nil
		}
	}
	return nil, nil
}
func (m *mockLabelRepo) Create(_ context.Context, label *domain.Label) error {
	if label.ID == "" {
		label.ID = id.New()
	}
	cp := *label
	m.store[label.ID] = &cp
	m.created = &cp
	return nil
}
func (m *mockLabelRepo) Rename(_ context.Context, label *domain.Label, newName string) error {
	m.renamedTo = newName
	if l, ok := m.store[label.ID]; ok {
		l.Name = newName
	}
	return nil
}
func (m *mockLabelRepo) Delete(_ context.Context, label *domain.Label) error {
	delete(m.store, label.ID)
	m.deletedIDs = append(m.deletedIDs, label.ID)
	return nil
}
