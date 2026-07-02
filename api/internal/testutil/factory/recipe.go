package factory

import "recipe-backend/internal/domain"

// DefaultOwnerID は NewRecipe の既定の所有者 ID(値に意味はないダミー UUID)。
const DefaultOwnerID = "00000000-0000-0000-0000-0000000000ff"

// RecipeOption は NewRecipe の生成オプション。
type RecipeOption func(*domain.Recipe)

// NewRecipe はテスト用の Recipe を生成する。
// デフォルトは「ID未設定・1人前・owner=1」。必要な属性だけオプションで上書きする。
func NewRecipe(opts ...RecipeOption) *domain.Recipe {
	r := &domain.Recipe{
		Title:    "テストレシピ",
		Servings: 1,
		OwnerID:  DefaultOwnerID,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// WithRecipeID は ID を指定する。
func WithRecipeID(id string) RecipeOption {
	return func(r *domain.Recipe) { r.ID = id }
}

// WithOwnerID は所有者 ID を指定する。
func WithOwnerID(id string) RecipeOption {
	return func(r *domain.Recipe) { r.OwnerID = id }
}

// WithTitle はタイトルを指定する。
func WithTitle(title string) RecipeOption {
	return func(r *domain.Recipe) { r.Title = title }
}

// WithSharedUsers は共有先ユーザーを指定する。
func WithSharedUsers(users ...domain.User) RecipeOption {
	return func(r *domain.Recipe) { r.SharedUsers = users }
}

// NewRecipeLabel はテスト用の RecipeLabel を生成する。
func NewRecipeLabel(name string) domain.RecipeLabel {
	return domain.RecipeLabel{Name: name}
}
