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
	// Update は username / email を更新する。
	Update(ctx context.Context, user *User) error
	// UpdatePassword は password_hash を更新する。
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	// Delete はユーザーと、そのユーザーが所有するレシピを削除する。
	// user 側の従属(labels / recipe_orders / recipe_archives / 共有)は FK CASCADE で消える。
	Delete(ctx context.Context, userID string) error
}

type LabelRepository interface {
	// FindAllForOwner は ownerID が管理するラベルを名前昇順で返す。
	FindAllForOwner(ctx context.Context, ownerID string) ([]Label, error)
	// FindByID は id のラベルを返す(無ければ nil)。所有チェック用。
	FindByID(ctx context.Context, id string) (*Label, error)
	// FindByOwnerAndName は (ownerID, name) のラベルを返す(無ければ nil)。重複チェック用。
	FindByOwnerAndName(ctx context.Context, ownerID, name string) (*Label, error)
	// Create はラベルを1件作る。
	Create(ctx context.Context, label *Label) error
	// Rename はラベル名を newName に変え、所有者のレシピの recipe_labels.name にも伝播する。
	Rename(ctx context.Context, label *Label, newName string) error
	// Delete はラベルを消し、所有者のレシピの recipe_labels からも同名を外す。
	Delete(ctx context.Context, label *Label) error
}

type RecipeRepository interface {
	// FindAllForUser は owner == userID または共有先に userID を含むレシピを、
	// そのユーザーの表示順(recipe_orders.position 昇順、未設定は末尾)で返す。
	// 各レシピの Archived には、この userID にとってのアーカイブ状態を詰める。
	FindAllForUser(ctx context.Context, userID string) ([]Recipe, error)
	FindByID(ctx context.Context, id string) (*Recipe, error)
	Create(ctx context.Context, recipe *Recipe) error
	Update(ctx context.Context, recipe *Recipe) error
	Delete(ctx context.Context, recipe *Recipe) error
	// Reorder は userID の表示順を recipeIDs の並び(先頭 = position 0)で保存する。
	// 他ユーザーの順序には影響しない。
	Reorder(ctx context.Context, userID string, recipeIDs []string) error
	// SetArchived は userID にとっての recipeID のアーカイブ状態を保存する。
	// archived=true で行を作り(冪等)、false で行を消す。他ユーザーには影響しない。
	SetArchived(ctx context.Context, userID, recipeID string, archived bool) error
	// IsArchived は userID にとって recipeID がアーカイブ済みかを返す。
	IsArchived(ctx context.Context, userID, recipeID string) (bool, error)
}
