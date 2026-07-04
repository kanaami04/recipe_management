package handler

import (
	"context"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/dto/request"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// --- テスト用 Echo（バリデータ付き）---

type testValidator struct{ v *validator.Validate }

func (tv *testValidator) Validate(i interface{}) error { return tv.v.Struct(i) }

func newTestEcho() *echo.Echo {
	e := echo.New()
	e.Validator = &testValidator{v: validator.New()}
	return e
}

// --- サービスのモック（関数フィールドで差し替え）---

type mockAuthService struct {
	loginFn    func(ctx context.Context, username, password string) (string, string, error)
	refreshFn  func(ctx context.Context, refresh string) (string, error)
	registerFn func(ctx context.Context, username, email, password string) (*domain.User, error)
}

func (m *mockAuthService) Login(ctx context.Context, u, p string) (string, string, error) {
	return m.loginFn(ctx, u, p)
}
func (m *mockAuthService) Refresh(ctx context.Context, r string) (string, error) {
	return m.refreshFn(ctx, r)
}
func (m *mockAuthService) Register(ctx context.Context, u, e, p string) (*domain.User, error) {
	return m.registerFn(ctx, u, e, p)
}

type mockRecipeService struct {
	listFn    func(ctx context.Context, userID string) ([]domain.Recipe, error)
	createFn  func(ctx context.Context, userID string, req request.RecipeRequest) (*domain.Recipe, error)
	updateFn  func(ctx context.Context, userID, recipeID string, req request.RecipeRequest) (*domain.Recipe, error)
	deleteFn  func(ctx context.Context, userID, recipeID string) error
	reorderFn func(ctx context.Context, userID string, recipeIDs []string) error
	archiveFn func(ctx context.Context, userID, recipeID string, archived bool) error
}

func (m *mockRecipeService) List(ctx context.Context, userID string) ([]domain.Recipe, error) {
	return m.listFn(ctx, userID)
}
func (m *mockRecipeService) Create(ctx context.Context, userID string, req request.RecipeRequest) (*domain.Recipe, error) {
	return m.createFn(ctx, userID, req)
}
func (m *mockRecipeService) Update(ctx context.Context, userID, recipeID string, req request.RecipeRequest) (*domain.Recipe, error) {
	return m.updateFn(ctx, userID, recipeID, req)
}
func (m *mockRecipeService) Delete(ctx context.Context, userID, recipeID string) error {
	return m.deleteFn(ctx, userID, recipeID)
}
func (m *mockRecipeService) Reorder(ctx context.Context, userID string, recipeIDs []string) error {
	return m.reorderFn(ctx, userID, recipeIDs)
}
func (m *mockRecipeService) SetArchived(ctx context.Context, userID, recipeID string, archived bool) error {
	return m.archiveFn(ctx, userID, recipeID, archived)
}

// --- UserService のモック ---

type mockUserService struct {
	getByIDFn  func(ctx context.Context, id string) (*domain.User, error)
	listFn     func(ctx context.Context) ([]domain.User, error)
	updateFn   func(ctx context.Context, userID, username, email string) (*domain.User, error)
	changePwFn func(ctx context.Context, userID, current, next string) error
	deleteFn   func(ctx context.Context, userID string) error
}

func (m *mockUserService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockUserService) List(ctx context.Context) ([]domain.User, error) {
	return m.listFn(ctx)
}
func (m *mockUserService) UpdateProfile(ctx context.Context, userID, username, email string) (*domain.User, error) {
	return m.updateFn(ctx, userID, username, email)
}
func (m *mockUserService) ChangePassword(ctx context.Context, userID, current, next string) error {
	return m.changePwFn(ctx, userID, current, next)
}
func (m *mockUserService) DeleteAccount(ctx context.Context, userID string) error {
	return m.deleteFn(ctx, userID)
}

// --- LabelService のモック ---

type mockLabelService struct {
	listFn   func(ctx context.Context, ownerID string) ([]domain.Label, error)
	createFn func(ctx context.Context, ownerID, name string) (*domain.Label, error)
	renameFn func(ctx context.Context, ownerID, id, name string) (*domain.Label, error)
	deleteFn func(ctx context.Context, ownerID, id string) error
}

func (m *mockLabelService) List(ctx context.Context, ownerID string) ([]domain.Label, error) {
	return m.listFn(ctx, ownerID)
}
func (m *mockLabelService) Create(ctx context.Context, ownerID, name string) (*domain.Label, error) {
	return m.createFn(ctx, ownerID, name)
}
func (m *mockLabelService) Rename(ctx context.Context, ownerID, id, name string) (*domain.Label, error) {
	return m.renameFn(ctx, ownerID, id, name)
}
func (m *mockLabelService) Delete(ctx context.Context, ownerID, id string) error {
	return m.deleteFn(ctx, ownerID, id)
}
