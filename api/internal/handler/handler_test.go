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
	getByIDFn       func(ctx context.Context, id string) (*domain.User, error)
	updateFn        func(ctx context.Context, userID, username string) (*domain.User, error)
	changeEmailFn   func(ctx context.Context, userID, email, password string) (*domain.User, error)
	changePwFn      func(ctx context.Context, userID, current, next string) error
	deleteFn        func(ctx context.Context, userID string) error
	createAvatarFn  func(ctx context.Context, userID, contentType string) (string, string, error)
	confirmAvatarFn func(ctx context.Context, userID, key string) (*domain.User, error)
	deleteAvatarFn  func(ctx context.Context, userID string) (*domain.User, error)
}

func (m *mockUserService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockUserService) UpdateProfile(ctx context.Context, userID, username string) (*domain.User, error) {
	return m.updateFn(ctx, userID, username)
}
func (m *mockUserService) ChangeEmail(ctx context.Context, userID, email, password string) (*domain.User, error) {
	return m.changeEmailFn(ctx, userID, email, password)
}
func (m *mockUserService) ChangePassword(ctx context.Context, userID, current, next string) error {
	return m.changePwFn(ctx, userID, current, next)
}
func (m *mockUserService) DeleteAccount(ctx context.Context, userID string) error {
	return m.deleteFn(ctx, userID)
}
func (m *mockUserService) CreateAvatarUploadURL(ctx context.Context, userID, contentType string) (string, string, error) {
	return m.createAvatarFn(ctx, userID, contentType)
}
func (m *mockUserService) ConfirmAvatar(ctx context.Context, userID, key string) (*domain.User, error) {
	return m.confirmAvatarFn(ctx, userID, key)
}
func (m *mockUserService) DeleteAvatar(ctx context.Context, userID string) (*domain.User, error) {
	return m.deleteAvatarFn(ctx, userID)
}

// --- AvatarStorage のモック ---

type mockAvatarStorage struct{}

func (mockAvatarStorage) PresignUpload(_ context.Context, key, _ string) (string, error) {
	return "https://example.com/upload/" + key, nil
}
func (mockAvatarStorage) Delete(_ context.Context, _ string) error { return nil }
func (mockAvatarStorage) PublicURL(key string) string              { return "https://example.com/public/" + key }

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

// --- ShoppingListService のモック ---

type mockShoppingListService struct {
	getFn          func(ctx context.Context, userID string) (*domain.ShoppingList, error)
	addItemFn      func(ctx context.Context, userID, listID, name string) (*domain.ShoppingList, error)
	setCheckedFn   func(ctx context.Context, userID, listID, itemID string, checked bool) (*domain.ShoppingList, error)
	deleteItemFn   func(ctx context.Context, userID, listID, itemID string) (*domain.ShoppingList, error)
	clearCheckedFn func(ctx context.Context, userID, listID string) (*domain.ShoppingList, error)
	reorderFn      func(ctx context.Context, userID, listID string, itemIDs []string) (*domain.ShoppingList, error)
}

func (m *mockShoppingListService) Get(ctx context.Context, userID string) (*domain.ShoppingList, error) {
	return m.getFn(ctx, userID)
}
func (m *mockShoppingListService) AddItem(ctx context.Context, userID, listID, name string) (*domain.ShoppingList, error) {
	return m.addItemFn(ctx, userID, listID, name)
}
func (m *mockShoppingListService) SetItemChecked(ctx context.Context, userID, listID, itemID string, checked bool) (*domain.ShoppingList, error) {
	return m.setCheckedFn(ctx, userID, listID, itemID, checked)
}
func (m *mockShoppingListService) DeleteItem(ctx context.Context, userID, listID, itemID string) (*domain.ShoppingList, error) {
	return m.deleteItemFn(ctx, userID, listID, itemID)
}
func (m *mockShoppingListService) ClearChecked(ctx context.Context, userID, listID string) (*domain.ShoppingList, error) {
	return m.clearCheckedFn(ctx, userID, listID)
}
func (m *mockShoppingListService) Reorder(ctx context.Context, userID, listID string, itemIDs []string) (*domain.ShoppingList, error) {
	return m.reorderFn(ctx, userID, listID, itemIDs)
}

// --- ShareGroupService のモック ---

type mockShareGroupService struct {
	getMineFn    func(ctx context.Context, userID string) (*domain.ShareGroup, error)
	createFn     func(ctx context.Context, userID, name string) (*domain.ShareGroup, error)
	joinFn       func(ctx context.Context, userID, code string, shareShoppingList bool) (*domain.ShareGroup, error)
	leaveFn      func(ctx context.Context, userID string) error
	removeFn     func(ctx context.Context, ownerID, targetUserID string) error
	regenerateFn func(ctx context.Context, ownerID string) (*domain.ShareGroup, error)
	sharingFn    func(ctx context.Context, userID string) (bool, error)
	setSharingFn func(ctx context.Context, userID string, share bool) error
}

func (m *mockShareGroupService) GetMine(ctx context.Context, userID string) (*domain.ShareGroup, error) {
	return m.getMineFn(ctx, userID)
}
func (m *mockShareGroupService) Create(ctx context.Context, userID, name string) (*domain.ShareGroup, error) {
	return m.createFn(ctx, userID, name)
}
func (m *mockShareGroupService) Join(ctx context.Context, userID, code string, shareShoppingList bool) (*domain.ShareGroup, error) {
	return m.joinFn(ctx, userID, code, shareShoppingList)
}
func (m *mockShareGroupService) Leave(ctx context.Context, userID string) error {
	return m.leaveFn(ctx, userID)
}
func (m *mockShareGroupService) RemoveMember(ctx context.Context, ownerID, targetUserID string) error {
	return m.removeFn(ctx, ownerID, targetUserID)
}
func (m *mockShareGroupService) RegenerateInviteCode(ctx context.Context, ownerID string) (*domain.ShareGroup, error) {
	return m.regenerateFn(ctx, ownerID)
}

// ShoppingListSharing は sharingFn 未設定時、既定として true(統合済み)を返す。
func (m *mockShareGroupService) ShoppingListSharing(ctx context.Context, userID string) (bool, error) {
	if m.sharingFn == nil {
		return true, nil
	}
	return m.sharingFn(ctx, userID)
}
func (m *mockShareGroupService) SetShoppingListSharing(ctx context.Context, userID string, share bool) error {
	return m.setSharingFn(ctx, userID, share)
}
