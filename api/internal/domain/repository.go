package domain

import (
	"context"
	"time"
)

// リポジトリ層のインターフェース。サービス層はこれらに依存する（依存方向を内側へ）。
// 第一引数の context はリクエスト由来の値（request_id 等）を下位層・GORM ログまで伝播させる。

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	Create(ctx context.Context, user *User) error
	// Update は username / email / email_verified を更新する。
	// メール変更時の「新アドレス保存 + 確認済みリセット」を 1 回の書き込みで原子的に行うため
	// email_verified も含める。
	Update(ctx context.Context, user *User) error
	// UpdatePassword は password_hash を更新する。
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	// UpdateAvatarKey は avatar_key を更新する。nil で未設定に戻す。
	UpdateAvatarKey(ctx context.Context, userID string, key *string) error
	// SetEmailVerified は email_verified を更新する。メール確認完了・メール変更時に使う。
	SetEmailVerified(ctx context.Context, userID string, verified bool) error
	// Delete はユーザーと、そのユーザーが所有するレシピを削除する。
	// user 側の従属(labels / recipe_orders / recipe_archives / 共有)は FK CASCADE で消える。
	Delete(ctx context.Context, userID string) error
}

// AvatarStorage はプロフィール画像の実体(S3 互換オブジェクトストレージ)を扱う抽象。
// GORM 経由の永続化ではなくオブジェクトストレージのため、他のリポジトリとは別インターフェースにする。
type AvatarStorage interface {
	// PresignUpload は key への PUT 用の署名付き URL を発行する。
	PresignUpload(ctx context.Context, key, contentType string) (url string, err error)
	// Delete は key のオブジェクトを削除する。
	Delete(ctx context.Context, key string) error
	// PublicURL は key を(相対または絶対の)公開 URL に変換する。ネットワークアクセスは行わない。
	PublicURL(key string) string
}

// Mailer はメール送信の抽象(本番は AWS SES)。GORM 経由の永続化ではなく外部送信のため、
// AvatarStorage と同様にリポジトリとは別インターフェースにする。link は完成した遷移先 URL。
type Mailer interface {
	// SendEmailVerification は確認リンク付きのメール確認メールを toEmail 宛に送る。
	SendEmailVerification(ctx context.Context, toEmail, link string) error
	// SendPasswordReset はリセットリンク付きのパスワードリセットメールを toEmail 宛に送る。
	SendPasswordReset(ctx context.Context, toEmail, link string) error
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
	// PruneRecipeState は userID が今は見られないレシピ(自分の所有でも、同じシェアグループの
	// メンバー所有でもないレシピ)に残った recipe_archives / recipe_orders 行を消す。
	// グループ脱退・除名・解散でレシピが見えなくなったとき、再共有で過去のアーカイブ状態・
	// 並び順が蘇らないよう掃除するのに使う(メンバー行の更新後に呼ぶ)。
	PruneRecipeState(ctx context.Context, userID string) error
}

type ShoppingListRepository interface {
	// FindByOwnerID は ownerID が所有する買い物リストを返す(無ければ nil)。
	// グループ所属時はグループ所有者の ID を渡すことで「グループの 1 リスト」を解決する。
	FindByOwnerID(ctx context.Context, ownerID string) (*ShoppingList, error)
	// FindByID は id のリストを返す(無ければ nil)。所有・認可チェック用。
	FindByID(ctx context.Context, id string) (*ShoppingList, error)
	// Create は空の買い物リストを 1 件作る。
	Create(ctx context.Context, list *ShoppingList) error
	// AddItem はリストに項目を 1 件追加する。
	AddItem(ctx context.Context, item *ShoppingListItem) error
	// AddItems はリストに複数項目をまとめて追加する(重複はマージせず別行で追加)。
	// position は既存の最大 + 1 から順に採番する。全項目が同じリストに属する前提。
	AddItems(ctx context.Context, items []*ShoppingListItem) error
	// SetItemChecked は項目のチェック状態を更新する。
	SetItemChecked(ctx context.Context, itemID string, checked bool) error
	// DeleteItem は項目を 1 件削除する。
	DeleteItem(ctx context.Context, itemID string) error
	// DeleteCheckedItems は listID のチェック済み項目をまとめて削除する。
	DeleteCheckedItems(ctx context.Context, listID string) error
	// Reorder は listID の項目の表示順を itemIDs の並び(先頭 = position 0)で保存する。
	Reorder(ctx context.Context, listID string, itemIDs []string) error
	// DeleteByOwnerID は ownerID が所有する買い物リストを物理削除する(無ければ何もしない)。
	// 項目は FK の ON DELETE CASCADE で一緒に消える。グループの共有リストへ統合するとき、
	// 統合前の個人リストを消すために使う。
	DeleteByOwnerID(ctx context.Context, ownerID string) error
}

type ShareGroupRepository interface {
	// Create はグループを作成し、所有者をメンバーに加える(トランザクション)。
	Create(ctx context.Context, group *ShareGroup) error
	// FindByUserID は userID が所属するグループを返す(無ければ nil)。Members を詰める。
	FindByUserID(ctx context.Context, userID string) (*ShareGroup, error)
	// FindByID は id のグループを返す(無ければ nil)。Members を詰める。
	FindByID(ctx context.Context, id string) (*ShareGroup, error)
	// FindByInviteCode は招待コードのグループを返す(無ければ nil)。有効期限は service で判定する。
	FindByInviteCode(ctx context.Context, code string) (*ShareGroup, error)
	// MemberIDs は userID と同じグループに属する全ユーザー ID(自分を含む)を返す。
	// どのグループにも属さないときは空スライスを返す。可視性・認可判定に使う。
	MemberIDs(ctx context.Context, userID string) ([]string, error)
	// AddMember は userID を groupID のメンバーに加える。shareShoppingList は参加時点の
	// 買い物リスト統合設定(true ならグループ所有者のリストを共同編集する)。
	AddMember(ctx context.Context, groupID, userID string, shareShoppingList bool) error
	// RemoveMember は userID を groupID のメンバーから外す。
	RemoveMember(ctx context.Context, groupID, userID string) error
	// FindMembership は userID 自身の ShareGroupMember 行を返す(どのグループにも属さなければ nil)。
	// ShareShoppingList など本人の所属設定を読むのに使う。
	FindMembership(ctx context.Context, userID string) (*ShareGroupMember, error)
	// SharingMemberIDs は groupID のメンバーのうち、買い物リストをグループに統合している
	// (ShareShoppingList = true)ユーザー ID を返す(所有者を含む)。
	SharingMemberIDs(ctx context.Context, groupID string) ([]string, error)
	// UpdateShareShoppingList は userID 自身の買い物リスト統合設定を更新する。
	UpdateShareShoppingList(ctx context.Context, userID string, share bool) error
	// UpdateInviteCode は招待コードと有効期限を差し替える(旧コードを失効させる)。
	UpdateInviteCode(ctx context.Context, groupID, code string, expiresAt time.Time) error
	// Delete はグループを解散する(メンバー行は FK CASCADE で消える)。
	Delete(ctx context.Context, groupID string) error
}
