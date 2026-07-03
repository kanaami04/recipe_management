package domain

import "context"

// リポジトリ層のインターフェース。サービス層はこれらに依存する（依存方向を内側へ）。
// 第一引数の context はリクエスト由来の値（request_id 等）を下位層・GORM ログまで伝播させる。

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	FindAll(ctx context.Context) ([]User, error)
	Create(ctx context.Context, user *User) error
}

type LabelRepository interface {
	// FindNamesForUser は userID が閲覧できる(所有 or 共有された)レシピの
	// ラベル名を重複なく昇順で返す。
	FindNamesForUser(ctx context.Context, userID string) ([]string, error)
}

type RecipeRepository interface {
	// FindAllForUser は owner == userID または共有先に userID を含むレシピを、
	// そのユーザーの表示順(recipe_orders.position 昇順、未設定は末尾)で返す。
	FindAllForUser(ctx context.Context, userID string) ([]Recipe, error)
	FindByID(ctx context.Context, id string) (*Recipe, error)
	Create(ctx context.Context, recipe *Recipe) error
	Update(ctx context.Context, recipe *Recipe) error
	Delete(ctx context.Context, recipe *Recipe) error
	// Reorder は userID の表示順を recipeIDs の並び(先頭 = position 0)で保存する。
	// 他ユーザーの順序には影響しない。
	Reorder(ctx context.Context, userID string, recipeIDs []string) error
}
