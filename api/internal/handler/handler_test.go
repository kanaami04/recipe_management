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
	registerFn func(ctx context.Context, username, email, password string) (*domain.ApplicationUser, error)
}

func (m *mockAuthService) Login(ctx context.Context, u, p string) (string, string, error) {
	return m.loginFn(ctx, u, p)
}
func (m *mockAuthService) Refresh(ctx context.Context, r string) (string, error) {
	return m.refreshFn(ctx, r)
}
func (m *mockAuthService) Register(ctx context.Context, u, e, p string) (*domain.ApplicationUser, error) {
	return m.registerFn(ctx, u, e, p)
}

type mockRecipeService struct {
	listFn   func(ctx context.Context, userID uint) ([]domain.Recipe, error)
	createFn func(ctx context.Context, userID uint, req request.RecipeRequest) (*domain.Recipe, error)
	updateFn func(ctx context.Context, userID, recipeID uint, req request.RecipeRequest) (*domain.Recipe, error)
	deleteFn func(ctx context.Context, userID, recipeID uint) error
}

func (m *mockRecipeService) List(ctx context.Context, userID uint) ([]domain.Recipe, error) {
	return m.listFn(ctx, userID)
}
func (m *mockRecipeService) Create(ctx context.Context, userID uint, req request.RecipeRequest) (*domain.Recipe, error) {
	return m.createFn(ctx, userID, req)
}
func (m *mockRecipeService) Update(ctx context.Context, userID, recipeID uint, req request.RecipeRequest) (*domain.Recipe, error) {
	return m.updateFn(ctx, userID, recipeID, req)
}
func (m *mockRecipeService) Delete(ctx context.Context, userID, recipeID uint) error {
	return m.deleteFn(ctx, userID, recipeID)
}
