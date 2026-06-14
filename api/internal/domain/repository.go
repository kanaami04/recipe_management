package domain

import "context"

// リポジトリ層のインターフェース。サービス層はこれらに依存する（依存方向を内側へ）。
// 第一引数の context はリクエスト由来の値（request_id 等）を下位層・GORM ログまで伝播させる。

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*ApplicationUser, error)
	FindByID(ctx context.Context, id uint) (*ApplicationUser, error)
	FindAll(ctx context.Context) ([]ApplicationUser, error)
	Create(ctx context.Context, user *ApplicationUser) error
}

type LabelRepository interface {
	FindAll(ctx context.Context) ([]RecipeLabel, error)
}

type RecipeRepository interface {
	// FindAllForUser は owner == userID または共有先に userID を含むレシピを返す。
	FindAllForUser(ctx context.Context, userID uint) ([]Recipe, error)
	FindByID(ctx context.Context, id uint) (*Recipe, error)
	Create(ctx context.Context, recipe *Recipe) error
	Update(ctx context.Context, recipe *Recipe) error
	Delete(ctx context.Context, recipe *Recipe) error

	// name による get-or-create（既存があれば再利用、無ければ作成）。
	GetOrCreateLabel(ctx context.Context, name string) (*RecipeLabel, error)
	GetOrCreateIngredient(ctx context.Context, name string) (*Ingredient, error)
	GetOrCreateSeasoning(ctx context.Context, name string) (*Seasoning, error)
}
